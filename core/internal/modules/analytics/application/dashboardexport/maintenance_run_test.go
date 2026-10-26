package dashboardexport

import (
	"context"
	"testing"
	"time"
)

func TestRecoveryCycleVisitsMaintenanceAfterOtherQueuesAndBindsLocalEvidence(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	localSHA := [32]byte{9}
	store := &maintenanceRecoveryStoreFixture{
		candidate: maintenanceCandidateFixture(maintenanceExpireKind, base), found: true,
		claim: MaintenanceUnknownClaim{Disposition: UnknownClaimed,
			CleanupGeneration: 1, DeleteAttemptGeneration: 0,
			CleanupLeaseUntil: base.Add(time.Minute), Mode: MaintenanceCleanupTerminal,
			TerminalState: StateExpired,
			StoredEvidence: MaintenanceResolutionEvidence{
				LocalStateSHA256: localSHA, LocalStateAt: base.Add(time.Second)}},
	}
	scan, cancel := context.WithCancel(context.Background())
	defer cancel()
	store.cancel = cancel
	runner := &RecoveryRunner{store: store, ids: workerIDs(),
		clock:     workerClockFixture{now: base.Add(5 * time.Second)},
		admission: dispatcherAdmissionFixture(true)}
	if err := runner.run(scan, context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.expiredScans != 1 || store.unknownScans != 1 ||
		store.maintenanceScans != 1 {
		t.Fatalf("cycle scans expired/unknown/maintenance = %d/%d/%d",
			store.expiredScans, store.unknownScans, store.maintenanceScans)
	}
	if store.resolved == nil || store.resolved.Kind != UnknownResolutionAbsent ||
		store.resolved.TerminalState != StateExpired ||
		store.resolved.Evidence.LocalStateSHA256 != localSHA ||
		store.resolved.SHA256 == [32]byte{} {
		t.Fatalf("terminal maintenance resolution = %#v", store.resolved)
	}
}

func TestMaintenanceUnknownExactDeleteDefersWithoutBlindReplay(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	draft := workerDraft()
	store := &maintenanceRecoveryStoreFixture{
		candidate: maintenanceCandidateFixture(maintenanceDeleteKind, base), found: true,
		claim: MaintenanceUnknownClaim{Disposition: UnknownClaimed,
			CleanupGeneration: 1, DeleteAttemptGeneration: 1,
			CleanupLeaseUntil: base.Add(time.Minute), Mode: MaintenanceCleanupArtifact,
			TerminalState: StateDeleted, Draft: &draft, ObjectVersion: "version-1",
			StoredEvidence: MaintenanceResolutionEvidence{
				LocalStateSHA256: [32]byte{8}, LocalStateAt: base.Add(time.Second)}},
	}
	artifacts := &artifactRuntimeFixture{
		reconciliation: Reconciliation{State: ReconciliationSealed,
			Receipt: UploadReceipt{ObjectVersion: "version-1",
				ProviderRequestID: "head-version", UploadedAt: base},
			EvidenceSHA256: [32]byte{2}, VerifiedAt: base.Add(time.Second)},
		deleteErr: ErrOutcomeUnknown,
	}
	runner := &RecoveryRunner{store: store, artifacts: artifacts, ids: workerIDs(),
		clock: workerClockFixture{now: base.Add(5 * time.Second)}}
	worked, err := runner.recoverMaintenanceUnknown(
		context.Background(), context.Background())
	if err != nil || !worked || !store.deferred || store.resolved != nil ||
		!artifacts.deleteCalled {
		t.Fatalf("maintenance defer = worked=%v deferred=%v resolved=%#v delete=%v err=%v",
			worked, store.deferred, store.resolved, artifacts.deleteCalled, err)
	}
}

func TestCompletedMaintenanceReplaysStoredDeleteEvidenceWithoutExternalCall(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	evidence := MaintenanceResolutionEvidence{ObjectKey: "analytics/export/object",
		ObjectVersion: "version-1", DeleteProviderRequestID: "delete-version",
		DeleteResponseSHA256: [32]byte{3}, AbsenceSHA256: [32]byte{4},
		DeletedAt: base, LocalStateSHA256: [32]byte{5},
		LocalStateAt: base.Add(time.Second)}
	store := &maintenanceRecoveryStoreFixture{
		candidate: maintenanceCandidateFixture(maintenanceDeleteKind, base), found: true,
		claim: MaintenanceUnknownClaim{Disposition: UnknownClaimed,
			CleanupGeneration: 1, DeleteAttemptGeneration: 1,
			CleanupLeaseUntil: base.Add(time.Minute), Mode: MaintenanceCleanupCompleted,
			TerminalState: StateDeleted, StoredEvidence: evidence},
	}
	runner := &RecoveryRunner{store: store, ids: workerIDs(),
		clock: workerClockFixture{now: base.Add(5 * time.Second)}}
	worked, err := runner.recoverMaintenanceUnknown(
		context.Background(), context.Background())
	if err != nil || !worked || store.resolved == nil ||
		store.resolved.Kind != UnknownResolutionDeleted ||
		store.resolved.Evidence.DeleteResponseSHA256 != evidence.DeleteResponseSHA256 {
		t.Fatalf("completed maintenance resolution = %#v, worked=%v err=%v",
			store.resolved, worked, err)
	}
}
