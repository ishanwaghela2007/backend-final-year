package kafka

import (
	"Auth/db"
	"Auth/internal/emailjob"
	"Auth/utils"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type ConsumerHandler struct {
	producer sarama.SyncProducer
	dlqTopic string
}

// NewConsumerHandler creates a new Kafka consumer handler
func NewConsumerHandler(producer sarama.SyncProducer, dlqTopic string) *ConsumerHandler {
	return &ConsumerHandler{
		producer: producer,
		dlqTopic: dlqTopic,
	}
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim consumes messages from Kafka and processes them
func (h *ConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var job emailjob.EmailJob

		if err := json.Unmarshal(msg.Value, &job); err != nil {
			log.Printf("[worker] invalid job payload: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}

		log.Printf("[worker] Processing job for email=%s subject=%s", job.To, job.Subject)

		// Retry sending email up to 3 times
		success := false
		for attempt := 1; attempt <= 3; attempt++ {
			if err := utils.SendWelcomeEmail(job.To, job.Subject); err != nil {
				log.Printf("[worker] Email send failed for %s (attempt %d): %v", job.To, attempt, err)
				time.Sleep(5 * time.Second)
			} else {
				success = true
				break
			}
		}

		if !success {
			log.Printf("[worker] Email failed after retries — sending to DLQ for %s", job.To)
			h.sendToDLQ(job)
			sess.MarkMessage(msg, "")
			continue
		}

		// ✅ Update user in Cassandra as verified
		if err := updateUserVerified(job.To); err != nil {
			log.Printf("[worker] Failed to update verification status for %s: %v", job.To, err)
		} else {
			log.Printf("[worker] ✅ User %s marked as verified in Cassandra", job.To)
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}

// sendToDLQ sends failed messages to Dead Letter Queue
func (h *ConsumerHandler) sendToDLQ(job emailjob.EmailJob) {
	data, _ := json.Marshal(job)
	msg := &sarama.ProducerMessage{
		Topic: h.dlqTopic,
		Value: sarama.ByteEncoder(data),
	}
	if _, _, err := h.producer.SendMessage(msg); err != nil {
		log.Printf("[DLQ] Failed to publish message: %v", err)
	} else {
		log.Printf("[DLQ] Message sent to DLQ topic: %s", h.dlqTopic)
	}
}

// updateUserVerified updates user status in Cassandra
func updateUserVerified(email string) error {
	query := `UPDATE users SET is_verified = true WHERE email = ?`

	if err := db.Session.Query(query, email).Exec(); err != nil {
		return err
	}
	return nil
}
