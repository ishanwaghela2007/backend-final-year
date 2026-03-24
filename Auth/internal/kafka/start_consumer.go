package kafka

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

func StartEmailConsumer(ctx context.Context, brokers []string, producer sarama.SyncProducer, dlqTopic string) error {
	group := "email-worker-group"
	consumer := NewConsumerHandler(producer, dlqTopic)

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return err
	}

	go func() {
		defer client.Close()
		for {
			if err := client.Consume(ctx, []string{"email_jobs"}, consumer); err != nil {
				log.Printf("[kafka-consumer] Error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}