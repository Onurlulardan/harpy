//go:build integration

package eventdelivery

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

const eventSchemaDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func integrationEventContracts(
	t *testing.T,
) (eventsv1.ProducerCatalog, eventsv1.ConsumerCatalog) {
	t.Helper()
	producer, err := eventsv1.NewProducerPolicy(
		"forms.entry.created", 1, "forms", eventSchemaDigest,
	)
	if err != nil {
		t.Fatal(err)
	}
	consumer, err := eventsv1.NewConsumerPolicy(
		"workflow.eventdelivery", 1, "workflow", "forms.entry.created", 1,
		"workflow.eventdelivery.v1", eventsv1.EffectDatabaseTransactional, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	producers, err := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{producer})
	if err != nil {
		t.Fatal(err)
	}
	consumers, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{consumer})
	if err != nil {
		t.Fatal(err)
	}
	return producers, consumers
}

func integrationEvent(
	t *testing.T,
	producers eventsv1.ProducerCatalog,
	eventID string,
	sequence int64,
) eventsv1.Envelope {
	t.Helper()
	tenantID := "22222222-2222-4222-8222-222222222222"
	payload := []byte(`{"value":1}`)
	value, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID: eventID, EventType: "forms.entry.created", EventVersion: 1,
		TenantID: &tenantID,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: tenantID, Domain: "forms",
			CutoverGeneration: 4, WriterOrigin: eventsv1.WriterV2,
			GroupID: "forms", GroupGeneration: 3,
		},
		AggregateType: "form.entry", AggregateID: "33333333-3333-4333-8333-333333333333",
		AggregateVersion: sequence, EventSequence: sequence,
		OccurredAt: "2026-07-15T10:00:00Z",
		ActorID:    "44444444-4444-4444-8444-444444444444", CorrelationID: "request-7",
		CausationID:    "55555555-5555-4555-8555-555555555555",
		IdempotencyKey: "event/id", BusinessEffectIdempotencyKey: "form/entry/id",
		PayloadSchemaSHA256: eventSchemaDigest,
		PayloadSHA256:       integrationDigest(payload), Payload: payload,
	}, producers)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func integrationDigest(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
