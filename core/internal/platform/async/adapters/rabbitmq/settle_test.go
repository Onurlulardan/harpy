package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestSettlementUsesAckOrSingleRejectOnlyAfterHandler(t *testing.T) {
	tests := []struct {
		name        string
		outcome     HandleOutcome
		handleErr   error
		wantAck     int
		wantReject  int
		wantRequeue bool
	}{
		{name: "commit", outcome: HandleCommitted, wantAck: 1},
		{name: "duplicate", outcome: HandleDuplicate, wantAck: 1},
		{name: "invalid parked", outcome: HandleOutcome("parked"), wantReject: 1},
		{name: "transient", outcome: HandleTransient, wantReject: 1, wantRequeue: true},
		{name: "permanent", outcome: HandlePermanent, wantReject: 1},
		{name: "handler error", handleErr: errors.New("processor failed"), wantReject: 1, wantRequeue: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			acknowledger := &fakeAcknowledger{}
			delivery, producers, consumers := validDelivery(t, acknowledger)
			handler := &fakeHandler{outcome: test.outcome, err: test.handleErr}
			err := settleDelivery(
				context.Background(), delivery, "workflow.projector", 1,
				producers, consumers, handler,
			)
			if err != nil || acknowledger.acks != test.wantAck ||
				acknowledger.rejects != test.wantReject ||
				acknowledger.requeue != test.wantRequeue || acknowledger.nacks != 0 ||
				handler.calls != 1 {
				t.Fatalf("unexpected settlement: err=%v ack=%+v handler=%+v", err, acknowledger, handler)
			}
		})
	}
}

func TestSettlementPermanentlyRejectsInvalidTransportBeforeHandler(t *testing.T) {
	acknowledger := &fakeAcknowledger{}
	delivery, producers, consumers := validDelivery(t, acknowledger)
	delivery.ContentType = "application/json"
	handler := &fakeHandler{outcome: HandleCommitted}
	if err := settleDelivery(
		context.Background(), delivery, "workflow.projector", 1,
		producers, consumers, handler,
	); err != nil {
		t.Fatal(err)
	}
	if acknowledger.rejects != 1 || acknowledger.requeue || handler.calls != 0 {
		t.Fatal("invalid transport reached the business handler")
	}
}

type fakeHandler struct {
	outcome HandleOutcome
	err     error
	calls   int
}

func (handler *fakeHandler) Handle(
	context.Context, eventsv1.Envelope, DeliveryMetadata,
) (HandleOutcome, error) {
	handler.calls++
	return handler.outcome, handler.err
}

type fakeAcknowledger struct {
	acks    int
	nacks   int
	rejects int
	requeue bool
}

func (value *fakeAcknowledger) Ack(uint64, bool) error {
	value.acks++
	return nil
}

func (value *fakeAcknowledger) Nack(uint64, bool, bool) error {
	value.nacks++
	return nil
}

func (value *fakeAcknowledger) Reject(_ uint64, requeue bool) error {
	value.rejects++
	value.requeue = requeue
	return nil
}

func validDelivery(
	t *testing.T, acknowledger amqp.Acknowledger,
) (amqp.Delivery, eventsv1.ProducerCatalog, eventsv1.ConsumerCatalog) {
	t.Helper()
	envelope := validEnvelope(t)
	body, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	producerPolicy, _ := eventsv1.NewProducerPolicy(
		"forms.entry.created", 1, "forms", envelope.PayloadSchemaSHA256,
	)
	producers, _ := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{producerPolicy})
	consumerPolicy, _ := eventsv1.NewConsumerPolicy(
		"workflow.projector", 1, "workflow", "forms.entry.created", 1,
		"workflow.projector.v1", eventsv1.EffectExternal, 8,
	)
	consumers, _ := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{consumerPolicy})
	return amqp.Delivery{
		Acknowledger: acknowledger, DeliveryTag: 7, Body: body,
		ContentType: eventsv1.BrokerContentType, DeliveryMode: amqp.Persistent,
		Headers: amqp.Table{"x-delivery-count": int64(2)}, Redelivered: true,
	}, producers, consumers
}
