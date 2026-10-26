package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func TestWriterRejectsCallerLocalJobCatalog(t *testing.T) {
	policy, err := jobsv1.NewKindPolicy(
		"rogue.task", 1, "workflow", jobsv1.EffectDatabaseTransactional, 3,
	)
	if err != nil {
		t.Fatal(err)
	}
	callerCatalog, _ := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	value, err := jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: "81000000-0000-4000-8000-000000000001",
		Kind:  "rogue.task", PayloadRevision: 1, TenantID: pointerString(consumerTenantA),
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: consumerTenantA,
			GroupID: "operations", GroupGeneration: 1,
			Domain: "workflow", CutoverGeneration: 1,
			WriterOrigin: jobsv1.WriterV2,
		},
		PayloadReference: "payload/rogue", PayloadSHA256: strings.Repeat("a", 64),
		RunAt: "2026-07-15T10:00:00Z", DedupeKey: "rogue/1", MaximumAttempts: 3,
	}, callerCatalog)
	if err != nil {
		t.Fatal(err)
	}
	empty, _ := jobsv1.NewCatalog(nil)
	repository := &repository{authority: writerCatalogAuthority(), jobs: empty}
	if _, err := repository.EnqueueJob(context.Background(), value); err == nil ||
		!strings.Contains(err.Error(), "not registered") {
		t.Fatalf("caller-local job catalog error = %v", err)
	}
}

func TestWriterRejectsCallerLocalEventCatalog(t *testing.T) {
	policy, err := eventsv1.NewProducerPolicy(
		"rogue.changed", 1, "workflow", strings.Repeat("b", 64),
	)
	if err != nil {
		t.Fatal(err)
	}
	callerCatalog, _ := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{policy})
	payload := []byte(`{"ok":true}`)
	digest := sha256.Sum256(payload)
	value, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID:   "82000000-0000-4000-8000-000000000001",
		EventType: "rogue.changed", EventVersion: 1,
		TenantID: pointerString(consumerTenantA),
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: consumerTenantA,
			GroupID: "operations", GroupGeneration: 1,
			Domain: "workflow", CutoverGeneration: 1,
			WriterOrigin: eventsv1.WriterV2,
		},
		AggregateType:    "workflow.instance",
		AggregateID:      "82100000-0000-4000-8000-000000000001",
		AggregateVersion: 1, EventSequence: 1,
		OccurredAt: "2026-07-15T10:00:00Z", ActorID: consumerActor,
		CorrelationID: "catalog:rogue.1", CausationID: consumerCause,
		IdempotencyKey: "event/rogue", BusinessEffectIdempotencyKey: "effect/rogue",
		PayloadSchemaSHA256: strings.Repeat("b", 64),
		PayloadSHA256:       hex.EncodeToString(digest[:]), Payload: payload,
	}, callerCatalog)
	if err != nil {
		t.Fatal(err)
	}
	empty, _ := eventsv1.NewProducerCatalog(nil)
	repository := &repository{authority: writerCatalogAuthority(), producers: empty}
	if _, err := repository.EnqueueEvent(context.Background(), value); err == nil {
		t.Fatalf("caller-local event catalog error = %v", err)
	}
}

func writerCatalogAuthority() authorityStamp {
	return authorityStamp{
		scopeID: uuid.MustParse(consumerTenantA), groupID: "operations",
		groupGeneration: 1, domain: "workflow", generation: 1,
	}
}

func pointerString(value string) *string { return &value }
