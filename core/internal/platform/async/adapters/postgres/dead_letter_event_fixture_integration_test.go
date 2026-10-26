//go:build integration

package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
)

func deadLetterEventRoot(
	t *testing.T,
	tenantID uuid.UUID,
	eventID uuid.UUID,
	aggregateID uuid.UUID,
	consumers eventsv1.ConsumerCatalog,
) deadletter.EventRoot {
	t.Helper()
	payload := []byte(`{"record_id":"42"}`)
	schemaSHA := deadLetterDigest([]byte("schema"))
	producer, err := eventsv1.NewProducerPolicy(
		"forms.entry.created", 1, "forms", schemaSHA,
	)
	if err != nil {
		t.Fatal(err)
	}
	producers, _ := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{producer})
	tenant := tenantID.String()
	envelope, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID: eventID.String(), EventType: "forms.entry.created", EventVersion: 1,
		TenantID: &tenant,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: tenant, Domain: "forms",
			CutoverGeneration: 1, WriterOrigin: eventsv1.WriterV2,
			GroupID: "billing_forms", GroupGeneration: 1,
		},
		AggregateType: "form.entry", AggregateID: aggregateID.String(),
		AggregateVersion: 1, EventSequence: 1,
		OccurredAt: "2026-07-15T10:00:00Z", ActorID: asyncActor,
		CorrelationID: "dead-letter-pg18", CausationID: asyncCausation,
		IdempotencyKey: "event/42", BusinessEffectIdempotencyKey: "form-entry/42",
		PayloadSchemaSHA256: schemaSHA, PayloadSHA256: deadLetterDigest(payload),
		Payload: payload,
	}, producers)
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	root, err := deadletter.NewEventRoot(
		document, "workflow.projector.v1",
		deadletter.DeliveryEvidence{
			Reason: deadletter.ReasonDeliveryLimit, DeliveryCount: 8,
		}, consumers,
	)
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func deadLetterDigest(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
