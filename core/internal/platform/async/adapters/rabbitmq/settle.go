package rabbitmq

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	amqp "github.com/rabbitmq/amqp091-go"
)

func settleDelivery(
	ctx context.Context,
	delivery amqp.Delivery,
	consumer string,
	consumerVersion int,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
	handler Handler,
) error {
	if handler == nil {
		return fmt.Errorf("event handler is unavailable")
	}
	if delivery.ContentType != eventsv1.BrokerContentType ||
		delivery.DeliveryMode != amqp.Persistent {
		return reject(delivery, false)
	}
	envelope, err := eventsv1.DecodeEnvelope(delivery.Body, producers)
	if err != nil {
		return reject(delivery, false)
	}
	policy, err := consumers.Resolve(
		consumer, consumerVersion, envelope.EventType, envelope.EventVersion,
	)
	if err != nil {
		return reject(delivery, false)
	}
	count, ok := deliveryCount(delivery.Headers)
	if !ok || count > policy.DeliveryLimit {
		return reject(delivery, false)
	}
	outcome, handleErr := handler.Handle(ctx, envelope, DeliveryMetadata{
		ConsumerPolicy: policy, DeliveryCount: count, Redelivered: delivery.Redelivered,
	})
	if handleErr != nil {
		return reject(delivery, true)
	}
	if !outcome.valid() {
		return reject(delivery, false)
	}
	switch outcome {
	case HandleCommitted, HandleDuplicate:
		return delivery.Ack(false)
	case HandleTransient:
		return reject(delivery, true)
	case HandlePermanent:
		return reject(delivery, false)
	default:
		return reject(delivery, false)
	}
}

func reject(delivery amqp.Delivery, requeue bool) error {
	return delivery.Reject(requeue)
}

func deliveryCount(headers amqp.Table) (int, bool) {
	value, exists := headers["x-delivery-count"]
	if !exists {
		return 0, true
	}
	var count int64
	switch typed := value.(type) {
	case int8:
		count = int64(typed)
	case int16:
		count = int64(typed)
	case int32:
		count = int64(typed)
	case int64:
		count = typed
	case int:
		count = int64(typed)
	default:
		return 0, false
	}
	if count < 0 || count > int64(^uint(0)>>1) {
		return 0, false
	}
	return int(count), true
}
