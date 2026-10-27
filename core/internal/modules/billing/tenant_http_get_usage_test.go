package billing

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestGetUsagePresentsCommittedAndReservedCurrent(t *testing.T) {
	fixture := newApplicationFixture(t)
	access := accessFor(t, fixture.access, billingapi.PermissionBillingUsageRead)
	view, err := fixture.application.GetUsage(
		context.Background(), billingapi.SubscriptionAccess{Access: access},
	)
	if err != nil || len(view.Meters) != 1 || view.Meters[0].Current != 9 ||
		view.Meters[0].Limit == nil || *view.Meters[0].Limit != 100 ||
		view.Meters[0].EntitlementKey != "forms.submissions" ||
		view.Meters[0].PeriodStart == nil || view.Meters[0].PeriodEnd == nil {
		t.Fatalf("usage view=%+v err=%v", view, err)
	}
}

func TestGetUsageRejectsCrossTenantResult(t *testing.T) {
	fixture := newApplicationFixture(t)
	access := accessFor(t, fixture.access, billingapi.PermissionBillingUsageRead)
	fixture.bindings.usage.TenantID = testV7(t)
	if _, err := fixture.application.GetUsage(
		context.Background(), billingapi.SubscriptionAccess{Access: access},
	); !errors.Is(err, billingapi.ErrUnavailable) {
		t.Fatalf("cross-tenant usage error = %v", err)
	}
}
