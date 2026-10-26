package analytics

import (
	"context"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func TestDashboardExportProviderRejectsForgedEffectIdentityBeforeWorker(t *testing.T) {
	t.Parallel()
	worker := &dashboardExportJobWorkerFixture{}
	provider, err := newDashboardExportJobProvider(DashboardExportDeleteJobKind,
		"delete", worker, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	request := dashboardExportEffectRequest(t, DashboardExportDeleteJobKind, "delete")
	versionFive, err := uuid.Parse(request.EffectID)
	if err != nil {
		t.Fatal(err)
	}
	versionFive[8] &= 0x3f
	if versionFive.Version() != uuid.Version(5) || versionFive.Variant() == uuid.RFC4122 {
		t.Fatalf("non-RFC UUIDv5 fixture is invalid: %s", versionFive)
	}
	for _, effectID := range []string{
		"019b1000-0000-7000-8000-000000000072", versionFive.String(),
	} {
		request.EffectID = effectID
		if result := provider.Execute(context.Background(), request); !result.IsUnknown() {
			t.Fatalf("forged effect owner %s result = %#v", effectID, result)
		}
	}
	if worker.calls != 0 {
		t.Fatalf("forged effect owners reached tenant worker: calls=%d", worker.calls)
	}
}

func TestMaintenanceRecoveryReceiptIDMatchesDefinitiveProvider(t *testing.T) {
	t.Parallel()
	effectID := uuid.NewSHA1(dashboardExportReceiptNamespace, []byte("effect-owner"))
	for _, test := range []struct{ kind, code string }{
		{platformasync.AnalyticsDashboardExportDeleteJobKind,
			"ANALYTICS_EXPORT_DELETED"},
		{platformasync.AnalyticsDashboardExportExpireJobKind,
			"ANALYTICS_EXPORT_EXPIRY_COMPLETE"},
	} {
		candidate := platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{
			Kind:             test.kind,
			TenantID:         uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			TenantSchema:     "tenant_analytics",
			ExportID:         uuid.MustParse("019b1000-0000-7000-8000-000000000061"),
			AuthorityGroupID: "analytics", AuthorityGroupGeneration: 1,
			CutoverGeneration: 1,
			JobID:             uuid.MustParse("019b1000-0000-7000-8000-000000000062"),
			AttemptID:         uuid.MustParse("019b1000-0000-7000-8000-000000000063"),
			AttemptNumber:     1, LeaseGeneration: 1,
			OriginalLeaseOwnerID: uuid.MustParse(
				"019b1000-0000-7000-8000-000000000064"),
			EffectID: effectID, EffectRequestSHA256: [32]byte{1},
			UnknownAt: time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC),
		}
		got, err := platformasync.AnalyticsDashboardExportMaintenanceReceiptID(candidate)
		want := dashboardExportReceiptID(effectsv1.Request{EffectID: effectID.String()},
			test.code)
		if err != nil || got.String() != want {
			t.Fatalf("maintenance receipt ID %s = %s, %v; want %s",
				test.kind, got, err, want)
		}
	}
}
