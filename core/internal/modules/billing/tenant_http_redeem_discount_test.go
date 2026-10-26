package billing

import (
	"context"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestRedeemDiscountUsesValidatedPinnedSubscription(t *testing.T) {
	fixture := newApplicationFixture(t)
	view, err := fixture.application.RedeemDiscount(
		context.Background(), billingapi.RedeemDiscountCommand{
			Access: fixture.access, IdempotencyKey: "redeem-1", Code: "launch",
		},
	)
	redemption := fixture.bindings.redeem
	if err != nil || view.ID != fixture.bindings.current.ID.String() ||
		redemption.SubscriptionID != fixture.bindings.eligibility.SubscriptionID ||
		redemption.OfferingVersionID != fixture.bindings.eligibility.OfferingVersionID ||
		redemption.PriceVersionID != fixture.bindings.eligibility.PriceVersionID ||
		redemption.Currency != "USD" || redemption.Code != "LAUNCH" ||
		fixture.bindings.resolveCode != "LAUNCH" ||
		redemption.SourceOperationID.Version() != 7 ||
		!commandReadModes(fixture.coordinator.modes) {
		t.Fatalf("view=%+v redemption=%+v modes=%v err=%v",
			view, redemption, fixture.coordinator.modes, err)
	}
}
