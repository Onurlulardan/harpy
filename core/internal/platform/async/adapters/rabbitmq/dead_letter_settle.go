package rabbitmq

import (
	"context"
	"fmt"
	"strings"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	amqp "github.com/rabbitmq/amqp091-go"
)

func settleDeadLetter(
	ctx context.Context,
	delivery amqp.Delivery,
	consumers eventsv1.ConsumerCatalog,
	ingest deadletter.IngestStore,
) error {
	if ingest == nil {
		return reject(delivery, true)
	}
	if delivery.ContentType != eventsv1.BrokerContentType ||
		delivery.DeliveryMode != amqp.Persistent {
		return reject(delivery, false)
	}
	queueName, ok := deadLetterQueue(delivery.RoutingKey)
	if !ok {
		return reject(delivery, false)
	}
	evidence, err := deadLetterEvidence(delivery.Headers)
	if err != nil {
		return reject(delivery, false)
	}
	root, err := deadletter.NewEventRoot(delivery.Body, queueName, evidence, consumers)
	if err != nil {
		return reject(delivery, false)
	}
	result, err := ingest.IngestEvent(ctx, root)
	if err != nil || result.RootID != root.RootID() || result.RecordedAt.IsZero() {
		return reject(delivery, true)
	}
	if err := delivery.Ack(false); err != nil {
		return fmt.Errorf("ack persisted RabbitMQ dead-letter: %w", err)
	}
	return nil
}

func deadLetterQueue(routingKey string) (string, bool) {
	const prefix = "consumer."
	if !strings.HasPrefix(routingKey, prefix) || len(routingKey) == len(prefix) {
		return "", false
	}
	queueName := strings.TrimPrefix(routingKey, prefix)
	return queueName, !strings.ContainsAny(queueName, " \t\r\n")
}
