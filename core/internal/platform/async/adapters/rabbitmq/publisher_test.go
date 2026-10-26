package rabbitmq

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestPublisherRequiresMandatoryPersistentConfirm(t *testing.T) {
	wire := newFakeWire(true)
	publisher, err := newPublisher(wire)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := publisher.Publish(context.Background(), PublishRequest{
		Envelope: validEnvelope(t), LeaseGeneration: 7,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.ConfirmSHA256) != 64 || !wire.mandatory || wire.immediate ||
		wire.exchange != eventsv1.BrokerExchangeName ||
		wire.routingKey != "forms.entry.created.v1" ||
		wire.message.DeliveryMode != amqp.Persistent ||
		wire.message.ContentType != eventsv1.BrokerContentType {
		t.Fatal("publisher reliability contract drifted")
	}
}

func TestPublisherRejectsReturnNackAndTimeout(t *testing.T) {
	tests := []struct {
		name string
		wire *fakeWire
		ctx  func() (context.Context, context.CancelFunc)
		code string
	}{
		{name: "return", wire: returningWire(), ctx: backgroundContext, code: FailureUnroutable},
		{name: "nack", wire: newFakeWire(false), ctx: backgroundContext, code: FailureNack},
		{name: "timeout", wire: blockingWire(), ctx: expiredContext, code: FailureTimeout},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publisher, _ := newPublisher(test.wire)
			ctx, cancel := test.ctx()
			defer cancel()
			_, err := publisher.Publish(ctx, PublishRequest{Envelope: validEnvelope(t), LeaseGeneration: 2})
			var failure *PublishFailure
			if !errors.As(err, &failure) || failure.Code() != test.code || len(failure.SHA256()) != 64 {
				t.Fatalf("unexpected publish failure: %v", err)
			}
		})
	}
}

type fakeConfirmation struct {
	ack   bool
	block bool
}

func (value fakeConfirmation) WaitContext(ctx context.Context) (bool, error) {
	if value.block {
		<-ctx.Done()
		return false, ctx.Err()
	}
	return value.ack, nil
}

type fakeWire struct {
	confirmation confirmation
	returns      chan amqp.Return
	closed       chan *amqp.Error
	exchange     string
	routingKey   string
	mandatory    bool
	immediate    bool
	message      amqp.Publishing
	closedByApp  bool
}

func newFakeWire(ack bool) *fakeWire {
	return &fakeWire{
		confirmation: fakeConfirmation{ack: ack},
		returns:      make(chan amqp.Return, 1), closed: make(chan *amqp.Error, 1),
	}
}

func returningWire() *fakeWire {
	wire := newFakeWire(true)
	wire.returns <- amqp.Return{MessageId: testEventID}
	return wire
}

func blockingWire() *fakeWire {
	wire := newFakeWire(false)
	wire.confirmation = fakeConfirmation{block: true}
	return wire
}

func (wire *fakeWire) Publish(
	_ context.Context, exchange string, routingKey string,
	mandatory bool, immediate bool, message amqp.Publishing,
) (confirmation, error) {
	wire.exchange, wire.routingKey = exchange, routingKey
	wire.mandatory, wire.immediate, wire.message = mandatory, immediate, message
	return wire.confirmation, nil
}

func (wire *fakeWire) Returns() <-chan amqp.Return { return wire.returns }
func (wire *fakeWire) Closed() <-chan *amqp.Error  { return wire.closed }
func (wire *fakeWire) Close() error {
	wire.closedByApp = true
	return nil
}

func backgroundContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func expiredContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Nanosecond)
}

const testEventID = "11111111-1111-4111-8111-111111111111"

func validEnvelope(t *testing.T) eventsv1.Envelope {
	t.Helper()
	schemaDigest := hexDigest("schema")
	policy, err := eventsv1.NewProducerPolicy("forms.entry.created", 1, "forms", schemaDigest)
	if err != nil {
		t.Fatal(err)
	}
	catalog, _ := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{policy})
	tenantID := "22222222-2222-4222-8222-222222222222"
	payload := []byte(`{"value":1}`)
	value, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID: testEventID, EventType: "forms.entry.created", EventVersion: 1,
		TenantID: &tenantID,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: tenantID, Domain: "forms",
			CutoverGeneration: 4, WriterOrigin: eventsv1.WriterV2,
			GroupID: "forms", GroupGeneration: 3,
		},
		AggregateType: "form.entry", AggregateID: "33333333-3333-4333-8333-333333333333",
		AggregateVersion: 5, EventSequence: 5, OccurredAt: "2026-07-15T10:00:00Z",
		ActorID: "44444444-4444-4444-8444-444444444444", CorrelationID: "request-7",
		CausationID:    "55555555-5555-4555-8555-555555555555",
		IdempotencyKey: "event/5", BusinessEffectIdempotencyKey: "form/entry/5",
		PayloadSchemaSHA256: schemaDigest, PayloadSHA256: digest(payload), Payload: payload,
	}, catalog)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func hexDigest(value string) string { return digest([]byte(value)) }
