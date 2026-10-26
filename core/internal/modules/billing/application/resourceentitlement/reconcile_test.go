package resourceentitlement

import (
	"context"
	"errors"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
)

func reconcileRequest() billingapi.MeteredResourceReconcileRequest {
	return billingapi.MeteredResourceReconcileRequest{
		Resource: billingapi.MeteredResourceAPIEndpoint,
		TenantID: mustV7(), ActorID: mustV7(), ProposedOperationID: mustV7(),
		OwnerFactID: mustV7(), IdempotencyKey: "reconcile-1",
		RequestSHA256: [32]byte{1},
	}
}

func reconciliationEvidence(
	request billingapi.MeteredResourceReconcileRequest,
	active, included bool,
) billingapi.MeteredResourceEvidence {
	return billingapi.MeteredResourceEvidence{
		Resource: request.Resource, TenantID: request.TenantID,
		OwnerFactID: request.OwnerFactID, Exists: active,
		Active: active, IncludedInOpening: included,
		EvidenceSHA256: [32]byte{7}, MeasuredAt: resourceNow,
	}
}

func TestReconcileAdoptsOnlyOpeningBackedActiveResource(t *testing.T) {
	store, usage := resourceFixture(t, 4, 5)
	request := reconcileRequest()
	evidence := reconciliationEvidence(request, true, true)
	commands, _ := NewReconcileCommands(store, usage, func() time.Time { return resourceNow })
	decision, err := commands.Reconcile(context.Background(), request, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Changed || decision.GaugeDelta != 0 ||
		len(usage.adjustCalls) != 0 || len(store.reconcileRecords) != 1 {
		t.Fatalf("unexpected adoption: %#v", decision)
	}
	change := store.reconcileRecords[0].Change
	if change == nil || change.Before != nil || change.After.State != CountedActive ||
		change.After.Revision != 1 {
		t.Fatalf("unexpected adoption change: %#v", change)
	}
}

func TestReconcileRejectsBlindActiveAdoption(t *testing.T) {
	store, usage := resourceFixture(t, 0, 5)
	request := reconcileRequest()
	evidence := reconciliationEvidence(request, true, false)
	commands, _ := NewReconcileCommands(store, usage, func() time.Time { return resourceNow })
	_, err := commands.Reconcile(context.Background(), request, evidence)
	if !errors.Is(err, billingapi.ErrMeteredResourceEvidence) ||
		len(usage.adjustCalls) != 0 || len(store.reconcileRecords) != 0 {
		t.Fatalf("blind adoption err=%v", err)
	}
}

func TestReconcileAuthoritativeReleaseAdjustsGauge(t *testing.T) {
	store, usage := resourceFixture(t, 2, 5)
	usage.allowed = true
	request := reconcileRequest()
	store.reconciliation.Projection = &ResourceProjection{
		Resource: request.Resource, OwnerFactID: request.OwnerFactID,
		State: CountedActive, Revision: 2, UpdatedAt: resourceNow,
	}
	evidence := reconciliationEvidence(request, false, false)
	commands, _ := NewReconcileCommands(store, usage, func() time.Time { return resourceNow })
	decision, err := commands.Reconcile(context.Background(), request, evidence)
	if err != nil || !decision.Changed || decision.GaugeDelta != -1 ||
		len(usage.adjustCalls) != 1 || usage.adjustCalls[0].Delta != -1 {
		t.Fatalf("release=%#v err=%v", decision, err)
	}
	change := store.reconcileRecords[0].Change
	if change.Before.Revision != 2 || change.After.State != CountedReleased ||
		change.After.Revision != 3 {
		t.Fatalf("unexpected release change: %#v", change)
	}
}

func TestReconcileRestoreFailsClosedWhenBuilderIsOff(t *testing.T) {
	store, usage := resourceFixture(t, 1, 5)
	store.feature = false
	request := reconcileRequest()
	store.reconciliation.Projection = &ResourceProjection{
		Resource: request.Resource, OwnerFactID: request.OwnerFactID,
		State: CountedReleased, Revision: 2, UpdatedAt: resourceNow,
	}
	evidence := reconciliationEvidence(request, true, false)
	commands, _ := NewReconcileCommands(store, usage, func() time.Time { return resourceNow })
	_, err := commands.Reconcile(context.Background(), request, evidence)
	if !errors.Is(err, billingapi.ErrMeteredResourceUnavailable) ||
		len(usage.adjustCalls) != 0 || len(store.reconcileRecords) != 0 {
		t.Fatalf("builder-off reconcile err=%v", err)
	}
}

func TestReconcileReplayPinsEvidenceAndCanonicalOperation(t *testing.T) {
	store, usage := resourceFixture(t, 0, 5)
	request := reconcileRequest()
	store.reconciliation.Projection = &ResourceProjection{
		Resource: request.Resource, OwnerFactID: request.OwnerFactID,
		State: CountedReleased, Revision: 1, UpdatedAt: resourceNow,
	}
	evidence := reconciliationEvidence(request, false, false)
	commands, _ := NewReconcileCommands(store, usage, func() time.Time { return resourceNow })
	first, err := commands.Reconcile(context.Background(), request, evidence)
	if err != nil {
		t.Fatal(err)
	}
	store.reconciliation.Stored = &StoredReconciliation{
		Decision: first, RecordedAt: resourceNow,
	}
	replayRequest := request
	replayRequest.ProposedOperationID = mustV7()
	replay, err := commands.Reconcile(context.Background(), replayRequest, evidence)
	if err != nil || !replay.Replay || replay.OperationID != first.OperationID {
		t.Fatalf("replay=%#v err=%v", replay, err)
	}
	evidence.EvidenceSHA256 = [32]byte{9}
	_, err = commands.Reconcile(context.Background(), replayRequest, evidence)
	if !errors.Is(err, billingapi.ErrMeteredResourceIdempotencyReused) {
		t.Fatalf("evidence reuse error=%v", err)
	}
}

func TestReconcileRejectsUnverifiedEvidenceBeforeStorage(t *testing.T) {
	store, usage := resourceFixture(t, 0, 5)
	request := reconcileRequest()
	evidence := reconciliationEvidence(request, false, false)
	evidence.MeasuredAt = resourceNow.In(time.FixedZone("wrong", 3600))
	commands, _ := NewReconcileCommands(store, usage, func() time.Time { return resourceNow })
	_, err := commands.Reconcile(context.Background(), request, evidence)
	if !errors.Is(err, billingapi.ErrMeteredResourceEvidence) || store.inspectCalls != 0 {
		t.Fatalf("unverified evidence err=%v inspect=%d", err, store.inspectCalls)
	}
}
