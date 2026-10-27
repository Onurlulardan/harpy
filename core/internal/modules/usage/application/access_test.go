package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

func TestNewAccessFailsClosedAcrossTenantAndControlScopes(t *testing.T) {
	tests := []struct {
		name  string
		input AccessInput
	}{
		{"tenant mismatch", AccessInput{ActorID: testActor, SubjectTenantID: testTenant,
			TargetTenantID: uuid.New(), EnvironmentID: testEnvironment,
			TokenVersion: 1, AuthorizationVersion: 1, AuthorityClock: 1,
			Scope: domain.TenantScope, Mode: ModeSaaS, Purpose: PurposeSaaSCustomer,
			Permission: PermissionTenantRead, Operation: OperationTenantOverview}},
		{"customer platform read", AccessInput{ActorID: testActor, SubjectTenantID: testTenant,
			EnvironmentID: testEnvironment, TokenVersion: 1, AuthorizationVersion: 1,
			AuthorityClock: 1, Scope: domain.PlatformScope, Mode: ModeSaaS,
			Purpose: PurposeSaaSCustomer, ControlPrimary: true,
			Permission: PermissionPlatformRead, Operation: OperationPlatformOverview}},
		{"control without primary fence", AccessInput{ActorID: testActor,
			SubjectTenantID: testTenant, EnvironmentID: testEnvironment,
			TokenVersion: 1, AuthorizationVersion: 1, AuthorityClock: 1,
			Scope: domain.PlatformScope, Mode: ModeSaaS, Purpose: PurposeSaaSControl,
			Permission: PermissionPlatformRead, Operation: OperationPlatformOverview}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewAccess(test.input); !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("expected forbidden, got %v", err)
			}
		})
	}
}

func TestReadUseCaseRequiresItsExactScopedOperation(t *testing.T) {
	service := testService()
	if _, err := service.Overview(context.Background(),
		tenantAccess(OperationTenantTrend), TimeRangeRequest{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("tenant cross-operation access accepted: %v", err)
	}
	if _, err := service.Trend(context.Background(),
		platformAccess(OperationPlatformOverview), TimeRangeRequest{},
		domain.GranularityHour); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("platform cross-operation access accepted: %v", err)
	}
}

func TestTimeRangeRejectsPartialAndFutureQueries(t *testing.T) {
	service := testService()
	start, future := testNow.Add(-time.Hour), testNow.Add(2*time.Minute)
	if _, err := service.timeRange(TimeRangeRequest{Start: &start}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("partial range accepted: %v", err)
	}
	if _, err := service.timeRange(TimeRangeRequest{Start: &start, End: &future}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("future range accepted: %v", err)
	}
}
