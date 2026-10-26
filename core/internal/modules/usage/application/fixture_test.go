package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

var (
	testActor       = uuid.MustParse("10000000-0000-4000-8000-000000000001")
	testTenant      = uuid.MustParse("20000000-0000-4000-8000-000000000002")
	testEnvironment = uuid.MustParse("30000000-0000-4000-8000-000000000003")
	testFact        = uuid.MustParse("01890f9e-7b6c-7cc7-98c4-dc0c0c07398f")
	testNow         = time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
)

type fixedClock struct{ value time.Time }

func (clock fixedClock) Now() time.Time { return clock.value }

type fixedIDs struct{ id uuid.UUID }

func (ids fixedIDs) NewV7() (uuid.UUID, error) { return ids.id, nil }

type stubStore struct {
	record func(domain.Observation, [32]byte) (RecordResult, error)
}

func (store *stubStore) Record(_ context.Context, value domain.Observation,
	digest [32]byte,
) (RecordResult, error) {
	if store.record != nil {
		return store.record(value, digest)
	}
	return RecordResult{FactID: uuid.UUID(value.ID)}, nil
}

func (*stubStore) Overview(context.Context, Access, domain.TimeRange, time.Time) (domain.Overview, error) {
	return domain.Overview{}, nil
}
func (*stubStore) ActiveUsers(context.Context, Access, time.Time) (domain.ActiveUsers, error) {
	return domain.ActiveUsers{}, nil
}
func (*stubStore) TopUsers(context.Context, Access, domain.RankedQuery) ([]domain.TopUser, error) {
	return nil, nil
}
func (*stubStore) UserSummary(context.Context, Access, string, domain.TimeRange) (domain.UserSummary, error) {
	return domain.UserSummary{}, nil
}
func (*stubStore) UserTimeline(context.Context, Access, domain.TimelineQuery) (domain.TimelinePage, error) {
	return domain.TimelinePage{}, nil
}
func (*stubStore) Features(context.Context, Access, domain.RankedQuery) ([]domain.FeatureUsage, error) {
	return nil, nil
}
func (*stubStore) Heatmap(context.Context, Access, domain.TimeRange) ([]domain.HeatmapCell, error) {
	return nil, nil
}
func (*stubStore) Trend(context.Context, Access, domain.TimeRange, domain.Granularity) ([]domain.TrendPoint, error) {
	return nil, nil
}

func testService() *Service {
	service, _ := NewService(&stubStore{}, fixedClock{testNow}, fixedIDs{testFact})
	return service
}

func tenantAccess(operation string) Access {
	access, _ := NewAccess(AccessInput{ActorID: testActor, SubjectTenantID: testTenant,
		TargetTenantID: testTenant, EnvironmentID: testEnvironment,
		TokenVersion: 1, AuthorizationVersion: 2, AuthorityClock: 3,
		Scope: domain.TenantScope, Mode: ModeSaaS, Purpose: PurposeSaaSCustomer,
		Permission: PermissionTenantRead, Operation: operation})
	return access
}

func platformAccess(operation string) Access {
	access, _ := NewAccess(AccessInput{ActorID: testActor, SubjectTenantID: testTenant,
		EnvironmentID: testEnvironment, TokenVersion: 1, AuthorizationVersion: 2,
		AuthorityClock: 3, Scope: domain.PlatformScope, Mode: ModeSaaS,
		Purpose: PurposeSaaSControl, ControlPrimary: true,
		Permission: PermissionPlatformRead, Operation: operation})
	return access
}
