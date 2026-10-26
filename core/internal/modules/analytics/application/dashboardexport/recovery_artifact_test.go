package dashboardexport

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUnknownCleanupPersistsExactOrderedDeleteEvidence(t *testing.T) {
	t.Parallel()
	uploaded := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	reconciledAt := uploaded.Add(time.Second)
	deletedAt := reconciledAt.Add(time.Second)
	runtime := &artifactRuntimeFixture{
		reconciliation: Reconciliation{State: ReconciliationSealed,
			Receipt: UploadReceipt{ObjectVersion: "version-1",
				ProviderRequestID: "lookup-1", UploadedAt: uploaded},
			EvidenceSHA256: [32]byte{1}, VerifiedAt: reconciledAt},
		deleteReceipt: DeleteReceipt{ProviderRequestID: "delete-1",
			DeleteResponseSHA256: [32]byte{2}, AbsenceSHA256: [32]byte{3},
			VerifiedAbsentAt: deletedAt},
	}
	draft := workerDraft()
	result := (&RecoveryRunner{artifacts: runtime}).cleanupUnknownArtifact(
		context.Background(), UnknownCleanupClaim{Draft: &draft})
	if result.err != nil || result.kind != UnknownResolutionDeleted {
		t.Fatalf("cleanup = %#v", result)
	}
	evidence := result.evidence
	if evidence.ObjectKey != draft.ObjectKey || evidence.ObjectVersion != "version-1" ||
		evidence.ReconcileProviderRequestID != "lookup-1" ||
		evidence.DeleteProviderRequestID != "delete-1" ||
		evidence.UploadedAt != uploaded || evidence.ReconciledAt != reconciledAt ||
		evidence.DeletedAt != deletedAt {
		t.Fatalf("evidence = %#v", evidence)
	}
}

func TestUnknownCleanupRejectsOutOfOrderEvidence(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		uploaded   time.Time
		reconciled time.Time
		deleted    time.Time
		lookupID   string
	}{
		{"missing lookup id", base, base.Add(time.Second), base.Add(2 * time.Second), ""},
		{"missing upload time", time.Time{}, base.Add(time.Second), base.Add(2 * time.Second), "lookup"},
		{"lookup before upload", base.Add(time.Second), base, base.Add(2 * time.Second), "lookup"},
		{"delete before lookup", base, base.Add(2 * time.Second), base.Add(time.Second), "lookup"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			runtime := &artifactRuntimeFixture{
				reconciliation: Reconciliation{State: ReconciliationSealed,
					Receipt: UploadReceipt{ObjectVersion: "v1",
						ProviderRequestID: test.lookupID, UploadedAt: test.uploaded},
					EvidenceSHA256: [32]byte{1}, VerifiedAt: test.reconciled},
				deleteReceipt: DeleteReceipt{ProviderRequestID: "delete",
					DeleteResponseSHA256: [32]byte{2}, AbsenceSHA256: [32]byte{3},
					VerifiedAbsentAt: test.deleted},
			}
			draft := workerDraft()
			result := (&RecoveryRunner{artifacts: runtime}).cleanupUnknownArtifact(
				context.Background(), UnknownCleanupClaim{Draft: &draft})
			if result.err == nil {
				t.Fatalf("accepted evidence: %#v", result.evidence)
			}
		})
	}
}

func TestUnknownResolutionDigestBindsEveryReceiptField(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	evidence := UnknownResolutionEvidence{ObjectKey: "key", ObjectVersion: "v1",
		ReconcileProviderRequestID: "lookup", DeleteProviderRequestID: "delete",
		ReconcileEvidenceSHA256: [32]byte{1}, DeleteResponseSHA256: [32]byte{2},
		AbsenceSHA256: [32]byte{3}, UploadedAt: base,
		ReconciledAt: base.Add(time.Second), DeletedAt: base.Add(2 * time.Second)}
	resolvedAt := base.Add(3 * time.Second)
	want := unknownResolutionDigest(UnknownResolutionDeleted, evidence, resolvedAt)
	mutations := []func(*UnknownResolutionEvidence){
		func(value *UnknownResolutionEvidence) { value.ObjectKey += "-drift" },
		func(value *UnknownResolutionEvidence) { value.ObjectVersion += "-drift" },
		func(value *UnknownResolutionEvidence) { value.ReconcileProviderRequestID += "-drift" },
		func(value *UnknownResolutionEvidence) { value.DeleteProviderRequestID += "-drift" },
		func(value *UnknownResolutionEvidence) { value.ReconcileEvidenceSHA256[0]++ },
		func(value *UnknownResolutionEvidence) { value.DeleteResponseSHA256[0]++ },
		func(value *UnknownResolutionEvidence) { value.AbsenceSHA256[0]++ },
		func(value *UnknownResolutionEvidence) { value.UploadedAt = value.UploadedAt.Add(time.Microsecond) },
		func(value *UnknownResolutionEvidence) { value.ReconciledAt = value.ReconciledAt.Add(time.Microsecond) },
		func(value *UnknownResolutionEvidence) { value.DeletedAt = value.DeletedAt.Add(time.Microsecond) },
	}
	for index, mutate := range mutations {
		changed := evidence
		mutate(&changed)
		if got := unknownResolutionDigest(UnknownResolutionDeleted, changed, resolvedAt); got == want {
			t.Fatalf("mutation %d did not change digest", index)
		}
	}
	if got := unknownResolutionDigest(UnknownResolutionDeleted, evidence,
		resolvedAt.Add(time.Microsecond)); got == want {
		t.Fatal("resolved timestamp was not bound")
	}
}

func TestTenantSchemaCannotStartWithUnderscore(t *testing.T) {
	t.Parallel()
	pin := rendererPin(FormatCSV)
	candidate := RenderCandidate{TenantID: pin.TenantID,
		TenantSchema: "_tenant", ExportID: pin.ExportID,
		JobID:             uuid.MustParse("019b1000-0000-7000-8000-000000000099"),
		AttemptGeneration: 1}
	if validateRenderCandidate(candidate) == nil {
		t.Fatal("candidate accepted a schema with a leading underscore")
	}
}
