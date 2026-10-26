package postgres

import (
	"errors"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func TestEventDecisionOnlyIssuesPermitForContiguousEvent(t *testing.T) {
	t.Parallel()
	producers, consumers := consumerCatalogs(t, eventsv1.EffectDatabaseTransactional)
	repository := consumerRepositoryFixture(producers, consumers)
	envelope := consumerEnvelope(
		t, producers, "68000000-0000-4000-8000-000000000001",
		"68100000-0000-4000-8000-000000000001", 3, consumerTenantA,
	)
	policy, _ := consumers.Resolve("projection.workflow", 1, "workflow.changed", 1)
	decision, err := repository.eventDecision(envelope, policy, "apply", 3)
	permit, permitted := decision.(EventEffectPermit)
	if err != nil || !permitted || permit.owner != repository.identity || permit.used == nil {
		t.Fatalf("contiguous event permit = %#v, %v", decision, err)
	}
	decision, err = repository.eventDecision(envelope, policy, "duplicate", 4)
	duplicate, duplicated := decision.(DuplicateEvent)
	if err != nil || !duplicated || duplicate.EventID() != envelope.EventID {
		t.Fatalf("duplicate event decision = %#v, %v", decision, err)
	}
}

func TestEventDecisionFailsClosedWithoutProducerEvidence(t *testing.T) {
	t.Parallel()
	producers, consumers := consumerCatalogs(t, eventsv1.EffectDatabaseTransactional)
	repository := consumerRepositoryFixture(producers, consumers)
	envelope := consumerEnvelope(
		t, producers, "68200000-0000-4000-8000-000000000001",
		"68300000-0000-4000-8000-000000000001", 1, consumerTenantA,
	)
	policy, _ := consumers.Resolve("projection.workflow", 1, "workflow.changed", 1)
	decision, err := repository.eventDecision(envelope, policy, "producer_missing", 1)
	if decision != nil || !errors.Is(err, ErrProducerEvidenceMissing) {
		t.Fatalf("missing producer decision = %#v, %v", decision, err)
	}
	decision, err = repository.eventDecision(envelope, policy, "unknown", 1)
	if decision != nil || !errors.Is(err, ErrConsumerInputInvalid) {
		t.Fatalf("unknown event decision = %#v, %v", decision, err)
	}
}
