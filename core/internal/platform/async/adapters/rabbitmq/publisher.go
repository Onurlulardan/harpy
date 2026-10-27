package rabbitmq

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	mutex  sync.Mutex
	wire   publishWire
	closed bool
}

func OpenPublisher(connection *amqp.Connection) (*Publisher, error) {
	wire, err := openWire(connection)
	if err != nil {
		return nil, fmt.Errorf("open RabbitMQ publisher: %w", err)
	}
	return newPublisher(wire)
}

func newPublisher(wire publishWire) (*Publisher, error) {
	if wire == nil || wire.Returns() == nil || wire.Closed() == nil {
		return nil, fmt.Errorf("RabbitMQ publisher dependencies are incomplete")
	}
	return &Publisher{wire: wire}, nil
}

func (publisher *Publisher) Publish(
	ctx context.Context, request PublishRequest,
) (PublishReceipt, error) {
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()
	if publisher.closed || request.LeaseGeneration < 1 {
		return PublishReceipt{}, newFailure(
			FailureClosed, request.Envelope.EventID, request.LeaseGeneration,
		)
	}
	routingKey, body, occurredAt, err := prepareMessage(request.Envelope)
	if err != nil {
		return PublishReceipt{}, err
	}
	waiter, err := publisher.wire.Publish(
		ctx, eventsv1.BrokerExchangeName, routingKey, true, false,
		amqp.Publishing{
			ContentType: eventsv1.BrokerContentType, DeliveryMode: amqp.Persistent,
			MessageId: request.Envelope.EventID, CorrelationId: request.Envelope.CorrelationID,
			Timestamp: occurredAt, Type: request.Envelope.EventType,
			AppId: "harpy-core-worker", Body: body,
		},
	)
	if err != nil || waiter == nil {
		publisher.closeBroken()
		return PublishReceipt{}, newFailure(FailurePublish, request.Envelope.EventID, request.LeaseGeneration)
	}
	acked, waitErr := waiter.WaitContext(ctx)
	if waitErr != nil {
		publisher.closeBroken()
		return PublishReceipt{}, newFailure(FailureTimeout, request.Envelope.EventID, request.LeaseGeneration)
	}
	if returned(publisher.wire.Returns(), request.Envelope.EventID) {
		return PublishReceipt{}, newFailure(FailureUnroutable, request.Envelope.EventID, request.LeaseGeneration)
	}
	if !acked {
		code := FailureNack
		select {
		case <-publisher.wire.Closed():
			code = FailureClosed
		default:
		}
		publisher.closeBroken()
		return PublishReceipt{}, newFailure(code, request.Envelope.EventID, request.LeaseGeneration)
	}
	return PublishReceipt{ConfirmSHA256: confirmDigest(request, routingKey, body)}, nil
}

func (publisher *Publisher) Close() error {
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()
	if publisher.closed {
		return nil
	}
	publisher.closed = true
	return publisher.wire.Close()
}

func (publisher *Publisher) closeBroken() {
	publisher.closed = true
	_ = publisher.wire.Close()
}

func prepareMessage(envelope eventsv1.Envelope) (string, []byte, time.Time, error) {
	routingKey, err := envelope.RoutingKey()
	if err != nil {
		return "", nil, time.Time{}, err
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return "", nil, time.Time{}, fmt.Errorf("marshal integration event: %w", err)
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, envelope.OccurredAt)
	if err != nil {
		return "", nil, time.Time{}, fmt.Errorf("parse integration event timestamp: %w", err)
	}
	return routingKey, body, occurredAt, nil
}

func returned(returns <-chan amqp.Return, eventID string) bool {
	select {
	case returned, open := <-returns:
		return !open || returned.MessageId == eventID
	default:
		return false
	}
}

func confirmDigest(request PublishRequest, routingKey string, body []byte) string {
	bodyDigest := sha256.Sum256(body)
	document := fmt.Sprintf(
		"rabbitmq-confirm.v1\n%s\n%d\n%s\n%s\n%s",
		request.Envelope.EventID, request.LeaseGeneration, eventsv1.BrokerExchangeName,
		routingKey, hex.EncodeToString(bodyDigest[:]),
	)
	digest := sha256.Sum256([]byte(document))
	return hex.EncodeToString(digest[:])
}
