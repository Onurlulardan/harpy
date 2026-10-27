package dashboardexport

import (
	"context"
	"testing"
	"time"
)

func TestMaintenanceMissingKeyStillDeletesPinnedExactVersion(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	runtime := &artifactRuntimeFixture{
		reconciliation: Reconciliation{State: ReconciliationMissing,
			Receipt:        UploadReceipt{ProviderRequestID: "head-current"},
			EvidenceSHA256: [32]byte{1}, VerifiedAt: now},
		deleteReceipt: DeleteReceipt{ProviderRequestID: "delete-version",
			DeleteResponseSHA256: [32]byte{2}, AbsenceSHA256: [32]byte{3},
			VerifiedAbsentAt: now.Add(time.Second)},
	}
	draft := workerDraft()
	claim := MaintenanceUnknownClaim{Draft: &draft, ObjectVersion: "version-pinned"}
	result := (&RecoveryRunner{artifacts: runtime}).cleanupMaintenanceArtifact(
		context.Background(), claim)
	if result.err != nil || result.kind != UnknownResolutionDeleted {
		t.Fatalf("cleanup = %#v", result)
	}
	if !runtime.deleteCalled || runtime.deleteVersion != "version-pinned" {
		t.Fatalf("DeleteExact called=%v version=%q",
			runtime.deleteCalled, runtime.deleteVersion)
	}
	if result.evidence.AbsenceSHA256 != runtime.deleteReceipt.AbsenceSHA256 ||
		result.evidence.ReconcileEvidenceSHA256 !=
			runtime.reconciliation.EvidenceSHA256 {
		t.Fatalf("resolution evidence = %#v", result.evidence)
	}
	if !result.evidence.UploadedAt.IsZero() ||
		result.evidence.ReconciledAt != now {
		t.Fatalf("missing reconciliation time shape = %#v", result.evidence)
	}
}

func TestMaintenanceSealedObjectCarriesUploadTimeIntoExactDelete(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	runtime := &artifactRuntimeFixture{
		reconciliation: Reconciliation{State: ReconciliationSealed,
			Receipt: UploadReceipt{ObjectVersion: "version-pinned",
				ProviderRequestID: "head-version", UploadedAt: now.Add(-time.Second)},
			EvidenceSHA256: [32]byte{1}, VerifiedAt: now},
		deleteReceipt: DeleteReceipt{ProviderRequestID: "delete-version",
			DeleteResponseSHA256: [32]byte{2}, AbsenceSHA256: [32]byte{3},
			VerifiedAbsentAt: now.Add(time.Second)},
	}
	draft := workerDraft()
	result := (&RecoveryRunner{artifacts: runtime}).cleanupMaintenanceArtifact(
		context.Background(), MaintenanceUnknownClaim{
			Draft: &draft, ObjectVersion: "version-pinned"})
	if result.err != nil || result.kind != UnknownResolutionDeleted {
		t.Fatalf("cleanup = %#v", result)
	}
	if result.evidence.UploadedAt != now.Add(-time.Second) ||
		result.evidence.ReconciledAt != now {
		t.Fatalf("sealed reconciliation time shape = %#v", result.evidence)
	}
}

func TestMaintenanceUnknownDeleteIsDeferredWithoutBlindReplay(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	runtime := &artifactRuntimeFixture{
		reconciliation: Reconciliation{State: ReconciliationSealed,
			Receipt: UploadReceipt{ObjectVersion: "v1", ProviderRequestID: "head",
				UploadedAt: now.Add(-time.Second)},
			EvidenceSHA256: [32]byte{1}, VerifiedAt: now},
		deleteErr: ErrOutcomeUnknown,
	}
	draft := workerDraft()
	result := (&RecoveryRunner{artifacts: runtime}).cleanupMaintenanceArtifact(
		context.Background(), MaintenanceUnknownClaim{Draft: &draft, ObjectVersion: "v1"})
	if result.err != ErrOutcomeUnknown || !runtime.deleteCalled {
		t.Fatalf("cleanup = %#v, deleteCalled=%v", result, runtime.deleteCalled)
	}
}
