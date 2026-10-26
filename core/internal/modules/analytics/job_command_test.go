package analytics

import (
	"context"
	"fmt"
	"testing"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/google/uuid"
)

type analyticsJobStoreFixture struct {
	warmed, verified, repaired bool
}

func (store *analyticsJobStoreFixture) WarmQuery(
	context.Context, uuid.UUID, uuid.UUID, [32]byte,
) error {
	store.warmed = true
	return nil
}
func (store *analyticsJobStoreFixture) VerifyRevalidated(
	context.Context, uuid.UUID, uuid.UUID, [32]byte,
) error {
	store.verified = true
	return nil
}
func (store *analyticsJobStoreFixture) ApplyPublicationRepair(
	context.Context, uuid.UUID, uint64, time.Time,
) error {
	store.repaired = true
	return nil
}

func TestAnalyticsDatabaseCommandsExecuteExactClaim(t *testing.T) {
	store := &analyticsJobStoreFixture{}
	manifest := fmt.Sprintf("%064x", 7)
	warmReference := fmt.Sprintf("analytics-cache-warm/%s/%s/%s",
		analyticsJobQuery, analyticsJobVersion, manifest)
	warm := &analyticsCacheWarmCommand{store: store}
	if _, err := warm.Execute(t.Context(), newAnalyticsJobClaim(
		asyncmanifestv1.AnalyticsCacheWarmJobKind, warmReference)); err != nil {
		t.Fatal(err)
	}
	revalidateReference := fmt.Sprintf("analytics-query-revalidate/%s/%s/%s",
		analyticsJobQuery, analyticsJobVersion, manifest)
	revalidate := &analyticsQueryRevalidateCommand{store: store}
	if _, err := revalidate.Execute(t.Context(), newAnalyticsJobClaim(
		asyncmanifestv1.AnalyticsQueryRevalidateJobKind, revalidateReference)); err != nil {
		t.Fatal(err)
	}
	repairReference := fmt.Sprintf("analytics-publication-repair/%s/4/active",
		analyticsJobVersion)
	repair := &analyticsPublicationRepairCommand{store: store,
		clock: func() time.Time {
			return time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
		}}
	if _, err := repair.Execute(t.Context(), newAnalyticsJobClaim(
		asyncmanifestv1.AnalyticsPublicationRepairJobKind, repairReference)); err != nil {
		t.Fatal(err)
	}
	if !store.warmed || !store.verified || !store.repaired {
		t.Fatalf("store calls=%+v", store)
	}
}
