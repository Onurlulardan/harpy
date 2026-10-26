package dashboard

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestCreateRejectsInvalidVisibility(t *testing.T) {
	_, err := (&Application{}).Create(context.Background(), platformtx.TenantScope{}, CreateRequest{Name: "X", Visibility: "public", LayoutDocument: []byte(`{}`), IdempotencyKey: "create-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("create error = %v", err)
	}
}

func TestCreateBoundRequiresExactBillingAdmissionForNewDashboard(t *testing.T) {
	port := &dashboardFixtureEntitlements{}
	store := &dashboardFixtureStore{create: func(
		_ context.Context, _ CreateMutation,
	) (Snapshot, bool, error) {
		if !port.called {
			t.Fatal("dashboard owner store called before Billing decision")
		}
		return dashboardFixtureSnapshot(false), false, nil
	}}
	application := dashboardCreateApplication(port)
	result, err := application.createBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), dashboardCreateRequest(), dashboardCreateLayout())
	if err != nil || result.DashboardID != dashboardFixtureID || !port.called {
		t.Fatalf("dashboard admission called=%v result=%+v err=%v",
			port.called, result, err)
	}
}

func TestCreateBoundPreservesCanonicalBillingAndOwnerReplay(t *testing.T) {
	port := &dashboardFixtureEntitlements{decision: dashboardFixtureMutationDecision(billingapi.MeteredResourceCreate, true)}
	application := dashboardCreateApplication(port)
	_, err := application.createBound(context.Background(), platformtx.TenantScope{},
		dashboardCreateStore(true), dashboardFixtureIdentity(), dashboardCreateRequest(),
		dashboardCreateLayout())
	if err != nil || !port.called {
		t.Fatalf("replay admission called=%v err=%v", port.called, err)
	}
}

func TestCreateBoundFailsClosedForBillingDenialMismatchAndUnavailable(t *testing.T) {
	tests := []struct {
		name string
		port *dashboardFixtureEntitlements
		want error
	}{
		{name: "nil", want: ErrUnavailable},
		{name: "denied", port: &dashboardFixtureEntitlements{decision: billingapi.MeteredResourceMutationDecision{
			Resource: billingapi.MeteredResourceAnalyticsDashboard,
			Mutation: billingapi.MeteredResourceCreate,
			TenantID: dashboardFixtureTenant, OperationID: dashboardFixtureOperation,
			OwnerFactID: dashboardFixtureID, FeatureEnabled: true,
			Current: 2, Limit: 2,
		}}, want: ErrEntitlement},
		{name: "mismatch", port: &dashboardFixtureEntitlements{decision: billingapi.MeteredResourceMutationDecision{
			Resource: billingapi.MeteredResourceAPIEndpoint,
			Mutation: billingapi.MeteredResourceCreate,
			TenantID: dashboardFixtureTenant, OperationID: dashboardFixtureOperation,
			OwnerFactID: dashboardFixtureID, FeatureEnabled: true,
			Allowed: true, GaugeDelta: 1, Unlimited: true,
		}}, want: ErrUnavailable},
		{name: "unavailable", port: &dashboardFixtureEntitlements{
			err: errors.New("billing unavailable")}, want: ErrUnavailable},
		{name: "billing conflict", port: &dashboardFixtureEntitlements{
			err: billingapi.ErrMeteredResourceStateConflict}, want: ErrConflict},
		{name: "billing invalid", port: &dashboardFixtureEntitlements{
			err: billingapi.ErrMeteredResourceRequest}, want: ErrInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ownerCalled := false
			store := &dashboardFixtureStore{create: func(
				_ context.Context, _ CreateMutation,
			) (Snapshot, bool, error) {
				ownerCalled = true
				return dashboardFixtureSnapshot(false), false, nil
			}}
			application := dashboardCreateApplication(test.port)
			_, err := application.createBound(context.Background(),
				platformtx.TenantScope{}, store,
				dashboardFixtureIdentity(), dashboardCreateRequest(), dashboardCreateLayout())
			if !errors.Is(err, test.want) {
				t.Fatalf("create error = %v, want %v", err, test.want)
			}
			if ownerCalled {
				t.Fatal("dashboard owner store called after Billing denied or failed closed")
			}
		})
	}
}

func TestCreateBoundPropagatesOwnerFailureAfterBillingAllow(t *testing.T) {
	port := &dashboardFixtureEntitlements{}
	store := &dashboardFixtureStore{create: func(
		_ context.Context, _ CreateMutation,
	) (Snapshot, bool, error) {
		if !port.called {
			t.Fatal("dashboard owner store called before Billing decision")
		}
		return Snapshot{}, false, ErrConflict
	}}
	application := dashboardCreateApplication(port)
	_, err := application.createBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), dashboardCreateRequest(), dashboardCreateLayout())
	if !errors.Is(err, ErrConflict) || !port.called {
		t.Fatalf("create owner failure called=%v err=%v", port.called, err)
	}
}

func dashboardCreateApplication(port billingapi.MeteredResourceMutationDecider) *Application {
	return &Application{entitlements: port,
		ids: &dashboardFixtureIDs{values: []uuid.UUID{
			dashboardFixtureOperation, dashboardFixtureID, dashboardFixtureVersion,
		}}, clock: dashboardFixtureClock{}}
}

func dashboardCreateStore(replay bool) *dashboardFixtureStore {
	return &dashboardFixtureStore{create: func(
		_ context.Context, _ CreateMutation,
	) (Snapshot, bool, error) {
		return dashboardFixtureSnapshot(false), replay, nil
	}}
}

func dashboardCreateRequest() CreateRequest {
	return CreateRequest{Name: "Operations", Visibility: "private",
		IdempotencyKey: "dashboard-create-1"}
}

func dashboardCreateLayout() Document {
	layout, _ := NewDocument([]byte(`{}`))
	return layout
}
