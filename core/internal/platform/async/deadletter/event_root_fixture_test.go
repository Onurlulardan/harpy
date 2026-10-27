package deadletter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func eventRootFixture(
	t *testing.T,
	subscriptionType string,
) ([]byte, eventsv1.ConsumerCatalog) {
	t.Helper()
	payload := []byte(`{"record_id":"42"}`)
	schemaSHA := testDigest([]byte("event schema"))
	producer, err := eventsv1.NewProducerPolicy(
		"forms.entry.created", 1, "forms", schemaSHA,
	)
	if err != nil {
		t.Fatal(err)
	}
	producers, _ := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{producer})
	tenantID := "11111111-1111-4111-8111-111111111111"
	envelope, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID:   "22222222-2222-4222-8222-222222222222",
		EventType: "forms.entry.created", EventVersion: 1, TenantID: &tenantID,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: tenantID, Domain: "forms",
			CutoverGeneration: 1, WriterOrigin: eventsv1.WriterV2,
			GroupID: "forms", GroupGeneration: 1,
		},
		AggregateType:    "form.entry",
		AggregateID:      "33333333-3333-4333-8333-333333333333",
		AggregateVersion: 1, EventSequence: 1,
		OccurredAt:                   "2026-07-15T10:00:00Z",
		ActorID:                      "44444444-4444-4444-8444-444444444444",
		CorrelationID:                "request-42",
		CausationID:                  "55555555-5555-4555-8555-555555555555",
		IdempotencyKey:               "event/42",
		BusinessEffectIdempotencyKey: "form-entry/42",
		PayloadSchemaSHA256:          schemaSHA, PayloadSHA256: testDigest(payload), Payload: payload,
	}, producers)
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	consumer, err := eventsv1.NewConsumerPolicy(
		"workflow.projector", 1, "workflow", subscriptionType, 1,
		"workflow.projector.v1", eventsv1.EffectDatabaseTransactional, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	consumers, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{consumer})
	if err != nil {
		t.Fatal(err)
	}
	return document, consumers
}

func testDigest(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
