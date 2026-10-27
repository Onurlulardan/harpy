package billing

import (
	"context"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestScheduleCancellationReturnsRereadCancellationView(t *testing.T) {
	fixture := newApplicationFixture(t)
	view, err := fixture.application.ScheduleCancellation(
		context.Background(), billingapi.ScheduleCancellationCommand{
			Access: fixture.access, IdempotencyKey: "cancel-1",
			ExpectedVersion: fixture.bindings.current.View.Version, Reason: "customer_request",
		},
	)
	if err != nil || !view.CancelAtPeriodEnd || view.Version != 4 ||
		fixture.bindings.cancel.Reason != "customer_request" ||
		fixture.bindings.cancel.ActorID != fixture.access.PrincipalID() ||
		!commandReadModes(fixture.coordinator.modes) {
		t.Fatalf("view=%+v cancel=%+v modes=%v err=%v",
			view, fixture.bindings.cancel, fixture.coordinator.modes, err)
	}
}
