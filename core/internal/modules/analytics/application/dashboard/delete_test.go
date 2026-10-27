package dashboard

import (
	"context"
	"errors"
	"testing"

	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestDeleteRejectsMissingRevision(t *testing.T) {
	_, err := (&Application{}).Delete(context.Background(), platformtx.TenantScope{}, DeleteRequest{DashboardID: dashboardFixtureID, IdempotencyKey: "delete-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("delete error = %v", err)
	}
}

func TestDeleteBoundRemainsAvailableWhenBuilderIsDisabled(t *testing.T) {
	storeCalled := false
	store := &dashboardFixtureStore{delete: func(
		context.Context, DeleteMutation,
	) (Snapshot, bool, error) {
		storeCalled = true
		result := dashboardFixtureSnapshot(false)
		result.Lifecycle = analyticsdashboard.LifecycleRetired
		result.RootRevision = 2
		return result, false, nil
	}}
	entitlements := &dashboardFixtureEntitlements{}
	application := &Application{
		entitlements: entitlements,
		ids: &dashboardFixtureIDs{values: []uuid.UUID{
			dashboardFixtureOperation,
		}},
		clock: dashboardFixtureClock{},
	}
	_, err := application.deleteBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), DeleteRequest{DashboardID: dashboardFixtureID,
			ExpectedRootRevision: 1, IdempotencyKey: "dashboard-delete-builder-off-1"})
	if err != nil || !storeCalled ||
		entitlements.request.Mutation != billingapi.MeteredResourceDelete {
		t.Fatalf("delete error=%v owner_called=%v mutation=%s",
			err, storeCalled, entitlements.request.Mutation)
	}
}

func TestDeleteBoundDoesNotWriteOwnerStateWhenBillingIsUnavailable(t *testing.T) {
	storeCalled := false
	store := &dashboardFixtureStore{delete: func(
		context.Context, DeleteMutation,
	) (Snapshot, bool, error) {
		storeCalled = true
		return Snapshot{}, false, nil
	}}
	application := &Application{entitlements: &dashboardFixtureEntitlements{
		err: errors.New("billing unavailable")},
		ids:   &dashboardFixtureIDs{values: []uuid.UUID{dashboardFixtureOperation}},
		clock: dashboardFixtureClock{}}
	_, err := application.deleteBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), DeleteRequest{DashboardID: dashboardFixtureID,
			ExpectedRootRevision: 1, IdempotencyKey: "dashboard-delete-2"})
	if !errors.Is(err, ErrUnavailable) || storeCalled {
		t.Fatalf("delete error=%v owner_called=%v", err, storeCalled)
	}
}
