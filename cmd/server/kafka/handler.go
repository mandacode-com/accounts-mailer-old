package kafkaserver

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaHandler interface {
	// HandleMessage processes a Kafka message.
	//
	// Parameters:
	//   - m: The Kafka message to process.
	// Returns:
	//   - error: An error if the message processing fails, nil otherwise.
	HandleMessage(ctx context.Context, m kafka.Message) error
}
