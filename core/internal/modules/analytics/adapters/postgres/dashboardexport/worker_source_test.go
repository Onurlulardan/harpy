package dashboardexport

import (
	"context"
	"errors"
	"testing"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type workerSourceFixture struct {
	candidate        platformasync.AnalyticsDashboardExportCandidate
	maintenance      platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate
	found            bool
	maintenanceFound bool
	tenant           uuid.UUID
}

func (source workerSourceFixture) NextAnalyticsDashboardExport(context.Context) (
	platformasync.AnalyticsDashboardExportCandidate, bool, error,
) {
	return source.candidate, source.found, nil
}

func (workerSourceFixture) NextAnalyticsDashboardExportUnknown(context.Context) (
	platformasync.AnalyticsDashboardExportUnknownCandidate, bool, error,
) {
	return platformasync.AnalyticsDashboardExportUnknownCandidate{}, false, nil
}

func (workerSourceFixture) NextAnalyticsDashboardExportExpired(context.Context) (
	platformasync.AnalyticsDashboardExportExpiredCandidate, bool, error,
) {
	return platformasync.AnalyticsDashboardExportExpiredCandidate{}, false, nil
}

func (source workerSourceFixture) NextAnalyticsDashboardExportMaintenanceUnknown(
	context.Context,
) (platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate, bool, error) {
	return source.maintenance, source.maintenanceFound, nil
}

func (source workerSourceFixture) AnalyticsDashboardExportSpoolTenant(
	context.Context, uuid.UUID,
) (uuid.UUID, bool, error) {
	return source.tenant, source.tenant != uuid.Nil, nil
}

func TestWorkerSourceMapsExactMaintenanceAuthority(t *testing.T) {
	t.Parallel()
	stored := platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{
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
		EffectID:            uuid.NewSHA1(uuid.NameSpaceOID, []byte("delete-effect")),
		EffectRequestSHA256: [32]byte{1},
		UnknownAt:           time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC),
	}
	store := &WorkerStore{source: workerSourceFixture{
		maintenance: stored, maintenanceFound: true}}
	candidate, found, err := store.NextMaintenanceUnknown(context.Background())
	if err != nil || !found || candidate.AuthorityGroupID != stored.AuthorityGroupID ||
		candidate.AuthorityGroupGeneration != stored.AuthorityGroupGeneration ||
		candidate.CutoverGeneration != stored.CutoverGeneration ||
		candidate.UnknownAt != stored.UnknownAt {
		t.Fatalf("NextMaintenanceUnknown() = %#v, %v, %v", candidate, found, err)
	}
}

func TestWorkerSourceAcceptsAdoptedUUIDv4Tenant(t *testing.T) {
	t.Parallel()
	stored := platformasync.AnalyticsDashboardExportCandidate{
		TenantID:          uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		TenantSchema:      "tenant_analytics",
		ExportID:          uuid.MustParse("019b1000-0000-7000-8000-000000000032"),
		JobID:             uuid.MustParse("019b1000-0000-7000-8000-000000000033"),
		AttemptGeneration: 1}
	store := &WorkerStore{source: workerSourceFixture{candidate: stored, found: true}}
	candidate, found, err := store.PeekRenderCandidate(context.Background())
	if err != nil || !found || candidate.TenantID != stored.TenantID ||
		candidate.TenantSchema != stored.TenantSchema {
		t.Fatalf("PeekRenderCandidate() = %#v, %v, %v", candidate, found, err)
	}
}

func TestMissingSpoolAuthorityIsInactiveWithoutTenantTransaction(t *testing.T) {
	t.Parallel()
	called := false
	store := &WorkerStore{source: workerSourceFixture{}, runInTenant: func(
		context.Context, uuid.UUID, func(context.Context, *workerBound) error,
	) error {
		called = true
		return nil
	}}
	active, err := store.ActiveSpoolReservation(context.Background(),
		exportapp.SpoolReservationIdentity{
			ExportID:          uuid.MustParse("019b1000-0000-7000-8000-000000000032"),
			AttemptGeneration: 1,
			LeaseOwnerID:      uuid.MustParse("019b1000-0000-7000-8000-000000000034")})
	if err != nil || active || called {
		t.Fatalf("ActiveSpoolReservation() = %v, %v; tenant=%v", active, err, called)
	}
}

type dispatchStateError string

func (state dispatchStateError) Error() string    { return string(state) }
func (state dispatchStateError) SQLState() string { return string(state) }

func TestDispatcherSQLStateMappingPreservesSpoolBoundary(t *testing.T) {
	t.Parallel()
	for state, want := range map[string]error{"P6046": exportapp.ErrDispatchCandidateChanged,
		"P6047": exportapp.ErrDispatchLeaseStale, "P6048": exportapp.ErrUnavailable,
		"P6049": exportapp.ErrUnavailable} {
		if got := mapDispatchError(dispatchStateError(state)); !errors.Is(got, want) {
			t.Fatalf("mapDispatchError(%s) = %v, want %v", state, got, want)
		}
	}
}
