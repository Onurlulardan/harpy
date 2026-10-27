package dashboardexport

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkerCarriesExactAttemptAndEffectEvidenceToReady(t *testing.T) {
	t.Parallel()
	pin, exportID, jobID, leaseOwner := workerClaimIdentity()
	claim := RenderClaim{Disposition: ClaimExecute, Pin: pin, JobID: jobID,
		LeaseOwner: leaseOwner, Generation: 1, State: StateRunning}
	store := &workerStoreFixture{claim: claim}
	uploadedAt := time.Date(2026, 8, 2, 11, 59, 0, 0, time.UTC)
	pending := &pendingFixture{receipt: UploadReceipt{ObjectVersion: "version-1",
		ProviderRequestID: "put-1", UploadedAt: uploadedAt}}
	worker, err := NewWorker(store, &workerAuthorizerFixture{}, workerRendererFixture{},
		&artifactRuntimeFixture{pending: pending}, workerClockFixture{now: dispatcherTime},
		workerIDs())
	if err != nil {
		t.Fatal(err)
	}
	attempt, spool := workerDispatchAttempt(pin, exportID, jobID, leaseOwner, claim)
	result, err := worker.ExecuteClaimedRender(context.Background(), attempt, spool)
	if err != nil || result.State != RenderExecutionReady || !spool.discarded {
		t.Fatalf("ExecuteClaimedRender() = %#v, %v", result, err)
	}
	want := renderAttemptIdentity(attempt)
	if store.effectWrite.Attempt != want || store.readyWrite.Attempt != want ||
		store.effectWrite.EffectID != store.readyWrite.EffectID ||
		store.effectWrite.EffectRequestSHA256 != store.readyWrite.EffectRequestSHA256 ||
		store.readyWrite.EffectReceiptID.Version() != 7 ||
		store.readyWrite.EffectResultSHA256 == [32]byte{} ||
		!store.readyWrite.Receipt.UploadedAt.Equal(uploadedAt) {
		t.Fatalf("effect=%#v ready=%#v want=%#v",
			store.effectWrite, store.readyWrite, want)
	}
}

func TestWorkerCancellationBeforeEffectFinishesCancelledAndDiscards(t *testing.T) {
	t.Parallel()
	pin, exportID, jobID, leaseOwner := workerClaimIdentity()
	claim := RenderClaim{Disposition: ClaimExecute, Pin: pin, JobID: jobID,
		LeaseOwner: leaseOwner, Generation: 1, State: StateRunning}
	store := &workerStoreFixture{claim: claim}
	worker, _ := NewWorker(store, &workerAuthorizerFixture{}, workerRendererFixture{},
		&artifactRuntimeFixture{pending: &pendingFixture{}},
		workerClockFixture{now: dispatcherTime}, workerIDs())
	attempt, spool := workerDispatchAttempt(pin, exportID, jobID, leaseOwner, claim)
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(ErrWorkerCancelled)
	result, err := worker.ExecuteClaimedRender(ctx, attempt, spool)
	if err != nil || result.State != RenderExecutionCancelled || !spool.discarded ||
		store.failureWrite.FailureClass != "cancelled" || store.effectStarted {
		t.Fatalf("cancel = %#v, %v; write=%#v", result, err, store.failureWrite)
	}
}

func TestWorkerForceCancellationAfterEffectRecordsUnknownAndRetains(t *testing.T) {
	t.Parallel()
	pin, exportID, jobID, leaseOwner := workerClaimIdentity()
	claim := RenderClaim{Disposition: ClaimExecute, Pin: pin, JobID: jobID,
		LeaseOwner: leaseOwner, Generation: 1, State: StateRunning}
	store := &workerStoreFixture{claim: claim}
	ctx, cancel := context.WithCancelCause(context.Background())
	pending := &pendingFixture{upload: func(context.Context) (UploadReceipt, error) {
		cancel(context.Canceled)
		return UploadReceipt{}, context.Canceled
	}}
	worker, _ := NewWorker(store, &workerAuthorizerFixture{}, workerRendererFixture{},
		&artifactRuntimeFixture{pending: pending},
		workerClockFixture{now: dispatcherTime}, workerIDs())
	attempt, spool := workerDispatchAttempt(pin, exportID, jobID, leaseOwner, claim)
	result, err := worker.ExecuteClaimedRender(ctx, attempt, spool)
	if err != nil || result.State != RenderExecutionOutcomeUnknown ||
		!store.unknown || spool.discarded || !store.effectStarted ||
		store.unknownWrite.Attempt != renderAttemptIdentity(attempt) ||
		store.unknownWrite.EffectID != store.effectWrite.EffectID ||
		store.unknownWrite.EffectRequestSHA256 != store.effectWrite.EffectRequestSHA256 {
		t.Fatalf("post-effect cancel = %#v, %v; unknown=%#v", result, err,
			store.unknownWrite)
	}
	if !errors.Is(context.Cause(ctx), context.Canceled) {
		t.Fatalf("context cause = %v", context.Cause(ctx))
	}
}

func TestWorkerExactStaleBeforeEffectRetainsSpoolAsLeaseLost(t *testing.T) {
	t.Parallel()
	pin, exportID, jobID, leaseOwner := workerClaimIdentity()
	claim := RenderClaim{Disposition: ClaimExecute, Pin: pin, JobID: jobID,
		LeaseOwner: leaseOwner, Generation: 1, State: StateRunning}
	store := &workerStoreFixture{claim: claim, markEffectErr: ErrDispatchLeaseStale}
	worker, _ := NewWorker(store, &workerAuthorizerFixture{}, workerRendererFixture{},
		&artifactRuntimeFixture{pending: &pendingFixture{}},
		workerClockFixture{now: dispatcherTime}, workerIDs())
	attempt, spool := workerDispatchAttempt(pin, exportID, jobID, leaseOwner, claim)
	result, err := worker.ExecuteClaimedRender(context.Background(), attempt, spool)
	if !errors.Is(err, ErrDispatchLeaseStale) ||
		result.State != RenderExecutionLeaseLost || spool.discarded || store.unknown {
		t.Fatalf("stale effect = %#v, %v; discarded=%v unknown=%v",
			result, err, spool.discarded, store.unknown)
	}
}
