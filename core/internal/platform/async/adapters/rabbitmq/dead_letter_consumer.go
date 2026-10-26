package rabbitmq

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DeadLetterConsumer struct {
	channel    *amqp.Channel
	deliveries <-chan amqp.Delivery
	consumers  eventsv1.ConsumerCatalog
	ingest     deadletter.IngestStore
}

func OpenDeadLetterConsumer(
	connection *amqp.Connection,
	consumers eventsv1.ConsumerCatalog,
	ingest deadletter.IngestStore,
) (*DeadLetterConsumer, error) {
	if connection == nil || connection.IsClosed() ||
		len(consumers.Policies()) == 0 || ingest == nil {
		return nil, fmt.Errorf("RabbitMQ dead-letter consumer dependencies are incomplete")
	}
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open RabbitMQ dead-letter channel: %w", err)
	}
	if err := channel.Qos(ConsumerPrefetch, 0, false); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("set RabbitMQ dead-letter prefetch: %w", err)
	}
	deliveries, err := channel.Consume(
		eventsv1.DeadLetterIngestQueue,
		"harpy.deadletters.ingest.v1",
		false, false, false, false, nil,
	)
	if err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("start RabbitMQ dead-letter consumer: %w", err)
	}
	return &DeadLetterConsumer{
		channel: channel, deliveries: deliveries, consumers: consumers, ingest: ingest,
	}, nil
}

func (consumer *DeadLetterConsumer) Run(ctx context.Context) error {
	if consumer == nil || consumer.channel == nil || consumer.deliveries == nil ||
		consumer.ingest == nil {
		return fmt.Errorf("RabbitMQ dead-letter consumer is unavailable")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, open := <-consumer.deliveries:
			if !open {
				return fmt.Errorf("RabbitMQ dead-letter delivery stream closed")
			}
			if err := settleDeadLetter(
				ctx, delivery, consumer.consumers, consumer.ingest,
			); err != nil {
				return fmt.Errorf("settle RabbitMQ dead-letter: %w", err)
			}
		}
	}
}

func (consumer *DeadLetterConsumer) Close() error {
	if consumer == nil || consumer.channel == nil {
		return nil
	}
	return consumer.channel.Close()
}
