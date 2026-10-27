package billing

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetCurrentSubscriptionUsesExactReadScope(t *testing.T) {
	fixture := newApplicationFixture(t)
	access := accessFor(t, fixture.access, billingapi.PermissionBillingRead)
	view, err := fixture.application.GetCurrentSubscription(
		context.Background(), billingapi.SubscriptionAccess{Access: access},
	)
	if err != nil || view.ID != fixture.bindings.current.ID.String() ||
		view.TenantID != access.TenantID().String() || view.PeriodStart == "" ||
		view.PeriodEnd == "" || view.CreatedAt == "" || view.UpdatedAt == "" ||
		view.EffectiveMinor != 9900 || view.EffectiveCurrency != "USD" ||
		view.Offering.State != "retired" || view.Offering.Public ||
		view.Offering.VersionID != view.OfferingVersionID || fixture.resolver.calls != 1 ||
		len(fixture.coordinator.modes) != 1 ||
		fixture.coordinator.modes[0] != platformtx.SnapshotQuery {
		t.Fatalf("view=%+v modes=%v resolve=%d err=%v",
			view, fixture.coordinator.modes, fixture.resolver.calls, err)
	}
}

func TestGetCurrentSubscriptionRejectsMixedTenantView(t *testing.T) {
	fixture := newApplicationFixture(t)
	access := accessFor(t, fixture.access, billingapi.PermissionBillingRead)
	fixture.bindings.current.View.TenantID = testV7(t).String()
	if _, err := fixture.application.GetCurrentSubscription(
		context.Background(), billingapi.SubscriptionAccess{Access: access},
	); !errors.Is(err, billingapi.ErrUnavailable) {
		t.Fatalf("mixed tenant subscription view error = %v", err)
	}
}

func TestGetCurrentSubscriptionRejectsWrongPermissionAndTenantResult(t *testing.T) {
	fixture := newApplicationFixture(t)
	if _, err := fixture.application.GetCurrentSubscription(
		context.Background(), billingapi.SubscriptionAccess{Access: fixture.access},
	); !errors.Is(err, billingapi.ErrForbidden) || fixture.resolver.calls != 0 {
		t.Fatalf("wrong permission err=%v resolve=%d", err, fixture.resolver.calls)
	}
	readAccess := accessFor(t, fixture.access, billingapi.PermissionBillingRead)
	fixture.bindings.current = currentSubscription(t, testV7(t))
	if _, err := fixture.application.GetCurrentSubscription(
		context.Background(), billingapi.SubscriptionAccess{Access: readAccess},
	); !errors.Is(err, billingapi.ErrUnavailable) {
		t.Fatalf("cross-tenant result error = %v", err)
	}
}
