package billing

import (
	"context"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestResumeCancellationReturnsRereadActiveView(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.bindings.current = revisedSubscription(fixture.bindings.current, true)
	view, err := fixture.application.ResumeCancellation(
		context.Background(), billingapi.ResumeCancellationCommand{
			Access: fixture.access, IdempotencyKey: "resume-1",
			ExpectedVersion: fixture.bindings.current.View.Version,
		},
	)
	if err != nil || view.CancelAtPeriodEnd || view.Version != 5 ||
		fixture.bindings.resume.SubscriptionID.String() != view.ID ||
		fixture.bindings.resume.SourceOperationID.Version() != 7 ||
		!commandReadModes(fixture.coordinator.modes) {
		t.Fatalf("view=%+v resume=%+v modes=%v err=%v",
			view, fixture.bindings.resume, fixture.coordinator.modes, err)
	}
}
