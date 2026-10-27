package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDeadLetterAckFollowsDurableIngestResult(t *testing.T) {
	delivery, acknowledger, consumers := deadLetterDelivery(t)
	sequence := []string{}
	acknowledger.sequence = &sequence
	ingest := &deadLetterIngestFake{sequence: &sequence}
	if err := settleDeadLetter(
		context.Background(), delivery, consumers, ingest,
	); err != nil {
		t.Fatal(err)
	}
	if len(sequence) != 2 || sequence[0] != "persist" || sequence[1] != "ack" ||
		!acknowledger.acked || acknowledger.rejected {
		t.Fatalf("dead-letter settlement order = %v, ack=%t reject=%t",
			sequence, acknowledger.acked, acknowledger.rejected)
	}
}

func TestDeadLetterStoreFailureRequeuesWithoutAck(t *testing.T) {
	delivery, acknowledger, consumers := deadLetterDelivery(t)
	ingest := &deadLetterIngestFake{err: errors.New("database unavailable")}
	if err := settleDeadLetter(
		context.Background(), delivery, consumers, ingest,
	); err != nil {
		t.Fatal(err)
	}
	if acknowledger.acked || !acknowledger.rejected || !acknowledger.requeue {
		t.Fatalf("failed persistence settlement ack=%t reject=%t requeue=%t",
			acknowledger.acked, acknowledger.rejected, acknowledger.requeue)
	}
}

func TestMalformedDeadLetterRoutesToDurableQuarantine(t *testing.T) {
	delivery, acknowledger, consumers := deadLetterDelivery(t)
	delivery.Body = []byte(`{"invalid":true}`)
	if err := settleDeadLetter(
		context.Background(), delivery, consumers, &deadLetterIngestFake{},
	); err != nil {
		t.Fatal(err)
	}
	if acknowledger.acked || !acknowledger.rejected || acknowledger.requeue {
		t.Fatal("malformed dead-letter was not terminally rejected to quarantine")
	}
}

func deadLetterDelivery(
	t *testing.T,
) (amqp.Delivery, *deadLetterAcknowledger, eventsv1.ConsumerCatalog) {
	t.Helper()
	envelope := validEnvelope(t)
	document, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	policy := consumerPolicy(t, "forms.entry.created", 1)
	consumers, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	acknowledger := &deadLetterAcknowledger{}
	delivery := amqp.Delivery{
		Acknowledger: acknowledger,
		ContentType:  eventsv1.BrokerContentType, DeliveryMode: amqp.Persistent,
		RoutingKey: "consumer.workflow.projector.v1", Body: document,
		Headers: amqp.Table{
			"x-first-death-reason": "rejected",
			"x-death":              []any{amqp.Table{"reason": "rejected", "count": int64(1)}},
		},
	}
	return delivery, acknowledger, consumers
}

type deadLetterIngestFake struct {
	sequence *[]string
	err      error
}

func (fake *deadLetterIngestFake) IngestEvent(
	_ context.Context,
	root deadletter.EventRoot,
) (deadletter.IngestResult, error) {
	if fake.sequence != nil {
		*fake.sequence = append(*fake.sequence, "persist")
	}
	if fake.err != nil {
		return deadletter.IngestResult{}, fake.err
	}
	return deadletter.IngestResult{
		RootID: root.RootID(), Created: true, RecordedAt: time.Now().UTC(),
	}, nil
}

type deadLetterAcknowledger struct {
	sequence *[]string
	acked    bool
	rejected bool
	requeue  bool
}

func (acknowledger *deadLetterAcknowledger) Ack(_ uint64, _ bool) error {
	acknowledger.acked = true
	if acknowledger.sequence != nil {
		*acknowledger.sequence = append(*acknowledger.sequence, "ack")
	}
	return nil
}

func (acknowledger *deadLetterAcknowledger) Nack(_ uint64, _ bool, _ bool) error {
	return errors.New("basic.nack is forbidden")
}

func (acknowledger *deadLetterAcknowledger) Reject(_ uint64, requeue bool) error {
	acknowledger.rejected, acknowledger.requeue = true, requeue
	if acknowledger.sequence != nil {
		*acknowledger.sequence = append(*acknowledger.sequence, "reject")
	}
	return nil
}
