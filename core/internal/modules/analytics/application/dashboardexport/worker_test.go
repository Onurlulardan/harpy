package dashboardexport

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWorkerReauthorizesImmediatelyBeforeArtifactEffect(t *testing.T) {
	t.Parallel()
	pin, exportID, jobID, leaseOwner := workerClaimIdentity()
	store := &workerStoreFixture{claim: RenderClaim{Disposition: ClaimExecute,
		Pin: pin, JobID: jobID, LeaseOwner: leaseOwner, Generation: 1, State: StateRunning}}
	authorizer := &workerAuthorizerFixture{failAt: 2}
	pending := &pendingFixture{}
	artifacts := &artifactRuntimeFixture{pending: pending}
	worker, err := NewWorker(store, authorizer, workerRendererFixture{}, artifacts,
		workerClockFixture{now: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}, workerIDs())
	if err != nil {
		t.Fatal(err)
	}
	attempt, spool := workerDispatchAttempt(pin, exportID, jobID, leaseOwner,
		store.claim)
	result, err := worker.ExecuteClaimedRender(context.Background(), attempt, spool)
	if err != nil || result.State != RenderExecutionDead ||
		authorizer.calls != 2 || pending.uploaded || store.effectStarted ||
		store.failedCode != "AUTHORIZATION_REVOKED" || !spool.discarded ||
		!artifacts.prepareReserved {
		t.Fatalf("ProcessClaimedRender() = %#v, %v; auth=%d upload=%v effect=%v failure=%s",
			result, err, authorizer.calls, pending.uploaded, store.effectStarted,
			store.failedCode)
	}
}

func TestWorkerRejectsLegacyRenderReconciliationClaim(t *testing.T) {
	t.Parallel()
	pin, exportID, jobID, leaseOwner := workerClaimIdentity()
	draft := workerDraft()
	store := &workerStoreFixture{claim: RenderClaim{Disposition: ClaimReconcile,
		Pin: pin, JobID: jobID, LeaseOwner: leaseOwner, Generation: 1,
		Draft: &draft, State: StateRunning}}
	artifacts := &artifactRuntimeFixture{pending: &pendingFixture{},
		reconciliation: Reconciliation{State: ReconciliationUnknown},
		reconcileErr:   ErrOutcomeUnknown}
	worker, err := NewWorker(store, &workerAuthorizerFixture{}, workerRendererFixture{},
		artifacts, workerClockFixture{now: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}, workerIDs())
	if err != nil {
		t.Fatal(err)
	}
	attempt, spool := workerDispatchAttempt(pin, exportID, jobID, leaseOwner,
		store.claim)
	result, err := worker.ExecuteClaimedRender(context.Background(), attempt, spool)
	if err != ErrInvalid || result.State != "" || store.unknown || spool.discarded {
		t.Fatalf("ExecuteClaimedRender() = %#v, %v; unknown=%v discarded=%v",
			result, err, store.unknown, spool.discarded)
	}
}

func TestWorkerPersistsUnknownUploadBeforeReturning(t *testing.T) {
	t.Parallel()
	pin, exportID, jobID, leaseOwner := workerClaimIdentity()
	store := &workerStoreFixture{claim: RenderClaim{Disposition: ClaimExecute,
		Pin: pin, JobID: jobID, LeaseOwner: leaseOwner, Generation: 1,
		State: StateRunning}}
	pending := &pendingFixture{err: ErrOutcomeUnknown}
	worker, err := NewWorker(store, &workerAuthorizerFixture{}, workerRendererFixture{},
		&artifactRuntimeFixture{pending: pending},
		workerClockFixture{now: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}, workerIDs())
	if err != nil {
		t.Fatal(err)
	}
	attempt, spool := workerDispatchAttempt(pin, exportID, jobID, leaseOwner,
		store.claim)
	result, err := worker.ExecuteClaimedRender(context.Background(), attempt, spool)
	if err != nil || result.State != RenderExecutionOutcomeUnknown ||
		!pending.uploaded || !store.effectStarted || !store.unknown || spool.discarded {
		t.Fatalf("ProcessClaimedRender() = %#v, %v; upload=%v effect=%v unknown=%v",
			result, err, pending.uploaded, store.effectStarted, store.unknown)
	}
}

func workerDispatchAttempt(
	pin Pin, exportID, jobID, leaseOwner uuid.UUID, claim RenderClaim,
) (RenderDispatchAttempt, *dispatcherSpool) {
	candidate := RenderCandidate{TenantID: pin.TenantID, ExportID: exportID,
		JobID: jobID, TenantSchema: "tenant_analytics", AttemptGeneration: 1}
	spool := &dispatcherSpool{identity: SpoolReservationIdentity{ExportID: exportID,
		AttemptGeneration: 1, LeaseOwnerID: leaseOwner}}
	return RenderDispatchAttempt{Candidate: candidate, Claim: claim,
		AttemptID: uuid.MustParse("019b1000-0000-7000-8000-00000000000c"),
		ClaimedAt: dispatcherTime, LeaseUntil: dispatcherTime.Add(time.Minute)}, spool
}

func TestDeleteWorkerFinishesPinnedCleanupAfterAuthorizationRevocation(t *testing.T) {
	t.Parallel()
	pin, exportID, jobID, leaseOwner := workerClaimIdentity()
	draft := workerDraft()
	store := &workerStoreFixture{deleteClaim: DeleteClaim{Pin: pin,
		Disposition: ClaimExecute, State: StateDeleting, Draft: &draft,
		ObjectVersion: "version-1", DeleteActorID: pin.ActorID,
		DeleteJobID: jobID, LeaseOwner: leaseOwner, DeleteGeneration: 1}}
	authorizer := &workerAuthorizerFixture{failAt: 1}
	artifacts := &artifactRuntimeFixture{}
	worker, err := NewWorker(store, authorizer, workerRendererFixture{}, artifacts,
		workerClockFixture{now: time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)}, workerIDs())
	if err != nil {
		t.Fatal(err)
	}
	result, err := worker.ProcessDelete(context.Background(), pin.TenantID,
		exportID, jobID, leaseOwner)
	if err != nil || result.State != StateDeleted || authorizer.calls != 0 ||
		!artifacts.deleteCalled || !store.deleted {
		t.Fatalf("ProcessDelete() = %#v, %v; auth=%d delete=%v complete=%v",
			result, err, authorizer.calls, artifacts.deleteCalled, store.deleted)
	}
}

func workerClaimIdentity() (Pin, uuid.UUID, uuid.UUID, uuid.UUID) {
	pin := rendererPin(FormatCSV)
	jobID := uuid.MustParse("019b1000-0000-7000-8000-00000000000a")
	leaseOwner := uuid.MustParse("019b1000-0000-7000-8000-00000000000b")
	return pin, pin.ExportID, jobID, leaseOwner
}
