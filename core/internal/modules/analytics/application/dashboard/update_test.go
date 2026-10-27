package dashboard

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestUpdateRejectsMissingRevision(t *testing.T) {
	_, err := (&Application{}).Update(context.Background(), platformtx.TenantScope{}, UpdateRequest{DashboardID: dashboardFixtureID, Name: "X", Visibility: "private", LayoutDocument: []byte(`{}`), IdempotencyKey: "update-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("update error = %v", err)
	}
}

func TestUpdateBoundDoesNotWriteOwnerStateWhenBuilderIsDisabled(t *testing.T) {
	storeCalled := false
	store := &dashboardFixtureStore{update: func(
		context.Context, UpdateMutation,
	) (Snapshot, bool, error) {
		storeCalled = true
		return Snapshot{}, false, nil
	}}
	decision := dashboardFixtureMutationDecision(billingapi.MeteredResourceUpdate, false)
	decision.FeatureEnabled, decision.Allowed = false, false
	application := &Application{entitlements: &dashboardFixtureEntitlements{decision: decision},
		ids:   &dashboardFixtureIDs{values: []uuid.UUID{dashboardFixtureOperation}},
		clock: dashboardFixtureClock{}}
	layout, _ := NewDocument([]byte(`{}`))
	_, err := application.updateBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), UpdateRequest{DashboardID: dashboardFixtureID,
			ExpectedRootRevision: 1, ExpectedVersionRevision: 1,
			Name: "Operations", Visibility: "private",
			IdempotencyKey: "dashboard-update-denied-1"}, layout)
	if !errors.Is(err, ErrEntitlement) || storeCalled {
		t.Fatalf("update error=%v owner_called=%v", err, storeCalled)
	}
}
