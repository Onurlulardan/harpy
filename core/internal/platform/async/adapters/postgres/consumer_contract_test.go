package postgres

import (
	"strings"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func TestEventReceiptBindingRejectsCrossTenantIdentity(t *testing.T) {
	t.Parallel()
	producers, consumers := consumerCatalogs(t, eventsv1.EffectDatabaseTransactional)
	repository := consumerRepositoryFixture(producers, consumers)
	envelope := consumerEnvelope(
		t, producers, "61000000-0000-4000-8000-000000000001",
		"62000000-0000-4000-8000-000000000001", 1, consumerTenantA,
	)
	policy, _ := consumers.Resolve("projection.workflow", 1, "workflow.changed", 1)
	permit := EventEffectPermit{owner: repository.identity, envelope: envelope, policy: policy}
	receipt := consumerEventReceipt(t, envelope, producers, consumers, nil)
	if !repository.matchesEventReceipt(permit, receipt) {
		t.Fatal("exact event receipt was rejected")
	}
	wrongTenant := consumerTenantB
	receipt.TenantID = &wrongTenant
	if repository.matchesEventReceipt(permit, receipt) {
		t.Fatal("tenant B receipt matched tenant A permit")
	}
}

func TestExternalReceiptMustBindDurableEventIdentity(t *testing.T) {
	t.Parallel()
	producers, consumers := consumerCatalogs(t, eventsv1.EffectExternal)
	repository := consumerRepositoryFixture(producers, consumers)
	envelope := consumerEnvelope(
		t, producers, "63000000-0000-4000-8000-000000000001",
		"64000000-0000-4000-8000-000000000001", 1, consumerTenantA,
	)
	policy, _ := consumers.Resolve("projection.workflow", 1, "workflow.changed", 1)
	permit := EventEffectPermit{owner: repository.identity, envelope: envelope, policy: policy}
	external := consumerExternalReceipt(t, envelope)
	eventReceipt := consumerEventReceipt(t, envelope, producers, consumers, &external.ReceiptID)
	if !repository.matchesExternalReceipt(permit, eventReceipt, external) {
		t.Fatal("exact external effect receipt was rejected")
	}
	external.CausationID = consumerCause
	if repository.matchesExternalReceipt(permit, eventReceipt, external) {
		t.Fatal("external receipt from another causation was accepted")
	}
}

func consumerEventReceipt(
	t *testing.T,
	envelope eventsv1.Envelope,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
	externalID *string,
) eventsv1.EffectReceipt {
	t.Helper()
	value, err := eventsv1.NewEffectReceipt(eventsv1.EffectReceipt{
		ReceiptID: "65000000-0000-4000-8000-000000000001",
		Consumer:  "projection.workflow", ConsumerVersion: 1,
		EventID: envelope.EventID, EventType: envelope.EventType,
		EventVersion: envelope.EventVersion, TenantID: envelope.TenantID,
		ProducerAuthority: envelope.ProducerAuthority,
		ConsumerAuthority: consumerEventAuthority(consumerTenantA),
		AggregateType:     envelope.AggregateType, AggregateID: envelope.AggregateID,
		ThroughSequence:         envelope.EventSequence,
		BusinessEffectKey:       envelope.BusinessEffectIdempotencyKey,
		ExternalEffectReceiptID: externalID, ResultReference: "result/event",
		ResultSHA256: strings.Repeat("d", 64), CorrelationID: envelope.CorrelationID,
		CausationID: envelope.EventID, OccurredAt: "2026-07-15T10:00:02Z",
	}, envelope, producers, consumers)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func consumerExternalReceipt(t *testing.T, envelope eventsv1.Envelope) effectsv1.Receipt {
	t.Helper()
	tenant := consumerTenantA
	request, err := effectsv1.NewRequest(
		"66000000-0000-4000-8000-000000000001", "workflow.event.effect", 1,
		&tenant, consumerEffectAuthority(tenant), "workflow-provider", "apply",
		envelope.BusinessEffectIdempotencyKey, envelope.CorrelationID, envelope.EventID,
		"payload/event", strings.Repeat("c", 64), "2026-07-15T10:00:00Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := effectsv1.NewSucceededReceipt(
		request, "67000000-0000-4000-8000-000000000001", "APPLIED", nil,
		"result/event", strings.Repeat("d", 64), "2026-07-15T10:00:01Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}
