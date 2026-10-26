//go:build integration

package deadletterdelivery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func managedConsumerCatalog(t *testing.T) eventsv1.ConsumerCatalog {
	t.Helper()
	policy, err := eventsv1.NewConsumerPolicy(
		"workflow.projector", 1, "workflow", "forms.entry.created", 1,
		"workflow.projector.v1", eventsv1.EffectExternal, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func managedEnvelopeDocument(t *testing.T, eventID string) []byte {
	t.Helper()
	payload := []byte(`{"value":1}`)
	schemaDigest := managedDigest([]byte("schema"))
	policy, err := eventsv1.NewProducerPolicy(
		"forms.entry.created", 1, "forms", schemaDigest,
	)
	if err != nil {
		t.Fatal(err)
	}
	producers, _ := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{policy})
	tenantID := "22222222-2222-4222-8222-222222222222"
	envelope, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID: eventID, EventType: "forms.entry.created", EventVersion: 1,
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
		PayloadSchemaSHA256: schemaDigest, PayloadSHA256: managedDigest(payload), Payload: payload,
	}, producers)
	if err != nil {
		t.Fatal(err)
	}
	document, err := envelope.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func managedDigest(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func publishManagedDeadLetter(
	t *testing.T,
	settings config.RabbitMQ,
	catalog eventsv1.ConsumerCatalog,
	document []byte,
) {
	t.Helper()
	connection := dialManagedRabbit(t, settings)
	defer connection.Close()
	if err := brokerrabbit.DeclareTopology(connection, catalog); err != nil {
		t.Fatal(err)
	}
	channel, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = channel.PublishWithContext(
		ctx, eventsv1.DeadLetterExchangeName, "consumer.workflow.projector.v1",
		true, false, amqp.Publishing{
			ContentType: eventsv1.BrokerContentType, DeliveryMode: amqp.Persistent,
			Body: document,
			Headers: amqp.Table{
				"x-first-death-reason": "rejected",
				"x-death":              []any{amqp.Table{"reason": "rejected", "count": int64(1)}},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func dialManagedRabbit(t *testing.T, settings config.RabbitMQ) *amqp.Connection {
	t.Helper()
	deadline := time.Now().Add(45 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		connection, err := brokerrabbit.Dial(ctx, settings)
		cancel()
		if err == nil {
			return connection
		}
		if time.Now().After(deadline) {
			t.Fatalf("RabbitMQ did not accept managed test connection: %v", err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
