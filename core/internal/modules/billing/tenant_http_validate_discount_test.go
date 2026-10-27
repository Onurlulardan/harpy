package billing

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestValidateDiscountReturnsTypedEligiblePolicy(t *testing.T) {
	fixture := newApplicationFixture(t)
	access := accessFor(t, fixture.access, billingapi.PermissionBillingRead)
	view, err := fixture.application.ValidateDiscount(
		context.Background(), billingapi.ValidateDiscountCommand{
			Access: access, Code: "launch",
			OfferingVersionID: fixture.bindings.current.OfferingVersionID.String(),
		},
	)
	if err != nil || view.ID != fixture.bindings.eligibility.DiscountID.String() ||
		view.Code != "LAUNCH" || view.Kind != "basis_points" ||
		view.BasisPoints == nil || *view.BasisPoints != 1500 || !view.AllOfferings ||
		len(view.OfferingIDs) != 0 || view.ValidStart == "" ||
		fixture.bindings.lookup.Code != "LAUNCH" {
		t.Fatalf("discount view=%+v err=%v", view, err)
	}
}

func TestValidateDiscountRejectsMixedIdentityView(t *testing.T) {
	fixture := newApplicationFixture(t)
	access := accessFor(t, fixture.access, billingapi.PermissionBillingRead)
	fixture.bindings.eligibility.View.ID = testV7(t).String()
	_, err := fixture.application.ValidateDiscount(
		context.Background(), billingapi.ValidateDiscountCommand{
			Access: access, Code: "launch",
			OfferingVersionID: fixture.bindings.current.OfferingVersionID.String(),
		},
	)
	if !errors.Is(err, billingapi.ErrUnavailable) {
		t.Fatalf("mixed discount identity error = %v", err)
	}
}

func TestValidateDiscountRejectsCrossTenantEligibility(t *testing.T) {
	fixture := newApplicationFixture(t)
	access := accessFor(t, fixture.access, billingapi.PermissionBillingRead)
	otherTenant := testV7(t)
	otherCurrent := currentSubscription(t, otherTenant)
	fixture.bindings.eligibility = eligibilitySnapshot(t, otherTenant, otherCurrent)
	_, err := fixture.application.ValidateDiscount(
		context.Background(), billingapi.ValidateDiscountCommand{
			Access: access, Code: "launch",
			OfferingVersionID: otherCurrent.OfferingVersionID.String(),
		},
	)
	if !errors.Is(err, billingapi.ErrUnavailable) {
		t.Fatalf("cross-tenant discount eligibility error = %v", err)
	}
}
