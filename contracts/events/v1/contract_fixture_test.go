package eventsv1

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

const (
	eventTenantID = "11111111-1111-4111-8111-111111111111"
	eventID       = "22222222-2222-4222-8222-222222222222"
	repairEventID = "33333333-3333-4333-8333-333333333333"
	receiptID     = "44444444-4444-4444-8444-444444444444"
	effectID      = "55555555-5555-4555-8555-555555555555"
	actorID       = "66666666-6666-4666-8666-666666666666"
	causationID   = "77777777-7777-4777-8777-777777777777"
	aggregateID   = "88888888-8888-4888-8888-888888888888"
	eventTime     = "2026-07-15T10:00:00Z"
)

func producerAuthority() Authority {
	return Authority{
		ScopeKind: ScopeTenant, ScopeID: eventTenantID, Domain: "forms", CutoverGeneration: 3,
		WriterOrigin: WriterV2, GroupID: "billing_forms", GroupGeneration: 4,
	}
}

func consumerAuthority() Authority {
	return Authority{
		ScopeKind: ScopeTenant, ScopeID: eventTenantID, Domain: "workflow", CutoverGeneration: 5,
		WriterOrigin: WriterV2, GroupID: "workflow", GroupGeneration: 6,
	}
}

func testConsumerCatalog(t *testing.T, effect EffectClass) ConsumerCatalog {
	t.Helper()
	policy, err := NewConsumerPolicy(
		"workflow.projector", 1, "workflow", "form.submitted", 1,
		"workflow.projector.form.submitted.v1", effect, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewConsumerCatalog([]ConsumerPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func testProducerCatalog(t *testing.T) ProducerCatalog {
	t.Helper()
	policy, err := NewProducerPolicy("form.submitted", 1, "forms", digest("c"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewProducerCatalog([]ProducerPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func envelopeInput() Envelope {
	tenant := eventTenantID
	return Envelope{
		EventID: eventID, EventType: "form.submitted", EventVersion: 1, TenantID: &tenant,
		ProducerAuthority: producerAuthority(), AggregateType: "form_submission", AggregateID: aggregateID,
		AggregateVersion: 7, EventSequence: 7, OccurredAt: eventTime, ActorID: actorID,
		CorrelationID: "request:forms.42", CausationID: causationID,
		IdempotencyKey: "form.submitted/42/7", BusinessEffectIdempotencyKey: "submission/42",
		PayloadSchemaSHA256: digest("c"), PayloadSHA256: digestBytes([]byte(`{"a":"value","z":1}`)),
		Payload: []byte(`{"z":1,"a":"value"}`),
	}
}

func effectReceiptInput() EffectReceipt {
	tenant := eventTenantID
	return EffectReceipt{
		ReceiptID: receiptID, Consumer: "workflow.projector", ConsumerVersion: 1, EventID: eventID,
		EventType: "form.submitted", EventVersion: 1, TenantID: &tenant,
		ProducerAuthority: producerAuthority(), ConsumerAuthority: consumerAuthority(),
		AggregateType: "form_submission", AggregateID: aggregateID,
		ThroughSequence: 7, BusinessEffectKey: "submission/42", ResultReference: "projection/submission/42",
		ResultSHA256: digest("a"), CorrelationID: "request:forms.42", CausationID: eventID,
		OccurredAt: "2026-07-15T10:00:01Z",
	}
}

func eventString(value string) *string { return &value }

func digest(character string) string {
	value := ""
	for len(value) < 64 {
		value += character
	}
	return value
}

func digestBytes(document []byte) string {
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:])
}
