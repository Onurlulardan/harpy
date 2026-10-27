package dashboardexport

import (
	"testing"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func TestMaintenanceResolutionArgumentsBindExactCandidateAndLocalFence(t *testing.T) {
	t.Parallel()
	claim, resolution := maintenanceResolutionFixture()
	platformCandidate := platformMaintenanceCandidate(claim.Candidate)
	effectReceiptID, err := platformasync.AnalyticsDashboardExportMaintenanceReceiptID(
		platformCandidate)
	if err != nil {
		t.Fatal(err)
	}
	effectReceiptSHA, err :=
		platformasync.AnalyticsDashboardExportMaintenanceReceiptSHA256(
			platformCandidate, effectReceiptID, string(resolution.TerminalState),
			resolution.ResolvedAt)
	if err != nil {
		t.Fatal(err)
	}
	arguments := maintenanceResolutionArguments(claim.Candidate.TenantSchema,
		claim, resolution, effectReceiptID, effectReceiptSHA)
	if len(arguments) != 38 {
		t.Fatalf("maintenance resolution argument count = %d", len(arguments))
	}
	checks := map[int]any{
		5:  claim.Candidate.AuthorityGroupID,
		6:  claim.Candidate.AuthorityGroupGeneration,
		7:  claim.Candidate.CutoverGeneration,
		15: claim.Candidate.UnknownAt,
		18: claim.DeleteAttemptGeneration,
		19: resolution.ReceiptID,
		20: effectReceiptID,
		26: resolution.Evidence.LocalStateAt,
		29: resolution.Evidence.ReconcileProviderRequestID,
		31: nil,
		37: resolution.ResolvedAt,
	}
	for index, want := range checks {
		if arguments[index] != want {
			t.Fatalf("argument[%d] = %#v, want %#v", index, arguments[index], want)
		}
	}
	if !validMaintenanceResolution(claim, resolution) {
		t.Fatal("exact missing-reconcile maintenance resolution was rejected")
	}
	resolution.Evidence.LocalStateSHA256[0]++
	if validMaintenanceResolution(claim, resolution) {
		t.Fatal("tampered local state evidence was accepted")
	}
}

func maintenanceResolutionFixture() (
	exportapp.MaintenanceUnknownClaim, exportapp.MaintenanceResolution,
) {
	base := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	candidate := exportapp.MaintenanceUnknownCandidate{
		Kind:             platformasync.AnalyticsDashboardExportDeleteJobKind,
		TenantID:         uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		TenantSchema:     "tenant_analytics",
		ExportID:         uuid.MustParse("019b1000-0000-7000-8000-000000000032"),
		AuthorityGroupID: "analytics", AuthorityGroupGeneration: 2,
		CutoverGeneration: 3,
		JobID:             uuid.MustParse("019b1000-0000-7000-8000-000000000033"),
		AttemptID:         uuid.MustParse("019b1000-0000-7000-8000-000000000034"),
		AttemptNumber:     2, LeaseGeneration: 2,
		OriginalLeaseOwnerID: uuid.MustParse(
			"019b1000-0000-7000-8000-000000000035"),
		EffectID:            uuid.NewSHA1(uuid.NameSpaceOID, []byte("maintenance-effect")),
		EffectRequestSHA256: [32]byte{1}, UnknownAt: base,
	}
	evidence := exportapp.MaintenanceResolutionEvidence{
		ObjectKey: "analytics/export", ObjectVersion: "version-1",
		ReconcileProviderRequestID: "head-current",
		DeleteProviderRequestID:    "delete-version",
		ReconcileEvidenceSHA256:    [32]byte{2}, DeleteResponseSHA256: [32]byte{3},
		AbsenceSHA256: [32]byte{4}, LocalStateSHA256: [32]byte{5},
		ReconciledAt: base.Add(2 * time.Second),
		DeletedAt:    base.Add(3 * time.Second), LocalStateAt: base.Add(time.Second),
	}
	claim := exportapp.MaintenanceUnknownClaim{Candidate: candidate,
		Disposition: exportapp.UnknownClaimed, CleanupGeneration: 1,
		DeleteAttemptGeneration: 2,
		CleanupLeaseOwnerID: uuid.MustParse(
			"019b1000-0000-7000-8000-000000000036"),
		CleanupLeaseUntil: base.Add(time.Minute),
		Mode:              exportapp.MaintenanceCleanupArtifact,
		TerminalState:     exportapp.StateDeleted, StoredEvidence: evidence,
	}
	resolution := exportapp.MaintenanceResolution{
		ReceiptID: uuid.MustParse("019b1000-0000-7000-8000-000000000037"),
		Kind:      exportapp.UnknownResolutionDeleted, SHA256: [32]byte{6},
		TerminalState: exportapp.StateDeleted, Evidence: evidence,
		ResolvedAt: base.Add(4 * time.Second),
	}
	return claim, resolution
}
