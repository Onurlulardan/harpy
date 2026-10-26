package rabbitmq

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	channel         *amqp.Channel
	deliveries      <-chan amqp.Delivery
	consumer        string
	consumerVersion int
	producers       eventsv1.ProducerCatalog
	consumers       eventsv1.ConsumerCatalog
	handler         Handler
}

func OpenConsumer(
	connection *amqp.Connection,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
	consumer string,
	consumerVersion int,
	handler Handler,
) (*Consumer, error) {
	if connection == nil || connection.IsClosed() || handler == nil {
		return nil, fmt.Errorf("RabbitMQ consumer dependencies are incomplete")
	}
	policy, err := consumers.ResolveIdentity(consumer, consumerVersion)
	if err != nil {
		return nil, err
	}
	tag, err := consumerTag(consumer, consumerVersion)
	if err != nil {
		return nil, err
	}
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open RabbitMQ consumer channel: %w", err)
	}
	if err := channel.Qos(ConsumerPrefetch, 0, false); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("set RabbitMQ consumer prefetch: %w", err)
	}
	deliveries, err := channel.Consume(
		policy.QueueName, tag, false, false, false, false, nil,
	)
	if err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("start RabbitMQ consumer: %w", err)
	}
	return &Consumer{
		channel: channel, deliveries: deliveries, consumer: consumer,
		consumerVersion: consumerVersion, producers: producers,
		consumers: consumers, handler: handler,
	}, nil
}

func (consumer *Consumer) Run(ctx context.Context) error {
	if consumer == nil || consumer.channel == nil || consumer.deliveries == nil {
		return fmt.Errorf("RabbitMQ consumer is unavailable")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, open := <-consumer.deliveries:
			if !open {
				return fmt.Errorf("RabbitMQ delivery stream closed")
			}
			if err := settleDelivery(
				ctx, delivery, consumer.consumer, consumer.consumerVersion,
				consumer.producers, consumer.consumers, consumer.handler,
			); err != nil {
				return fmt.Errorf("settle RabbitMQ delivery: %w", err)
			}
		}
	}
}

func (consumer *Consumer) Close() error {
	if consumer == nil || consumer.channel == nil {
		return nil
	}
	return consumer.channel.Close()
}
