package kafka

import (
	"Auth/internal/emailjob"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

var Producer sarama.SyncProducer

// InitProducer initializes a Kafka SyncProducer
func InitProducer(brokers []string) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return err
	}
	Producer = producer
	log.Println("✅ Kafka producer initialized")
	return nil
}

// CloseProducer safely closes the Kafka producer
func CloseProducer() {
	if Producer != nil {
		if err := Producer.Close(); err != nil {
			log.Printf("⚠️ Failed to close Kafka producer: %v", err)
		}
	}
}

// PublishEmailJob publishes an email job to Kafka
func PublishEmailJob(job emailjob.EmailJob) error {
	if Producer == nil {
		return ErrProducerNotReady
	}

	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "email_jobs", // 👈 must match consumer topic
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = Producer.SendMessage(msg)
	if err != nil {
		log.Printf("❌ Failed to send Kafka message: %v", err)
		return err
	}

	log.Printf("[producer] Published email job to topic email_jobs for %s", job.To)
	return nil
}
// ErrProducerNotReady is returned when the producer isn't initialized
var ErrProducerNotReady = sarama.ConfigurationError("Kafka producer not initialized")
