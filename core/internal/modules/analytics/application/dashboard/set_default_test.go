package dashboard

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestSetDefaultRequiresIdempotencyKey(t *testing.T) {
	err := (&Application{}).SetDefault(context.Background(), platformtx.TenantScope{}, SetDefaultRequest{DashboardID: dashboardFixtureID, Surface: "home"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("set default error = %v", err)
	}
}

func TestSetDefaultDoesNotWriteOwnerStateWhenBuilderIsDisabled(t *testing.T) {
	storeCalled := false
	store := &dashboardFixtureStore{setDefault: func(
		context.Context, DefaultMutation,
	) (bool, error) {
		storeCalled = true
		return false, nil
	}}
	decision := dashboardFixtureMutationDecision(billingapi.MeteredResourceUpdate, false)
	decision.FeatureEnabled, decision.Allowed = false, false
	application := &Application{
		entitlements: &dashboardFixtureEntitlements{decision: decision},
		ids: &dashboardFixtureIDs{values: []uuid.UUID{
			dashboardFixtureOperation,
		}},
		clock: dashboardFixtureClock{},
	}
	err := application.setDefaultBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), SetDefaultRequest{
			DashboardID: dashboardFixtureID, Surface: "home",
			IdempotencyKey: "dashboard-default-denied-1",
		})
	if !errors.Is(err, ErrEntitlement) || storeCalled {
		t.Fatalf("set default error=%v owner_called=%v", err, storeCalled)
	}
}
