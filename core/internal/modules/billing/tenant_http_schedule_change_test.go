package billing

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestScheduleChangeCommandsThenRereadsSameAuthority(t *testing.T) {
	fixture := newApplicationFixture(t)
	targetOffering, targetPrice := testV7(t), testV7(t)
	view, err := fixture.application.ScheduleChange(context.Background(),
		billingapi.ScheduleChangeCommand{
			Access: fixture.access, IdempotencyKey: "change-1",
			SubscriptionID:          fixture.bindings.current.ID.String(),
			TargetOfferingVersionID: targetOffering.String(),
			TargetPriceVersionID:    targetPrice.String(),
			ExpectedVersion:         fixture.bindings.current.View.Version,
		})
	if err != nil || fixture.bindings.change.SubscriptionID.String() != view.ID ||
		fixture.bindings.change.TargetOfferingVersionID != targetOffering ||
		fixture.bindings.change.TargetPriceVersionID != targetPrice ||
		fixture.bindings.change.SourceOperationID.Version() != 7 || view.Version != 4 ||
		fixture.resolver.calls != 1 || !commandReadModes(fixture.coordinator.modes) {
		t.Fatalf("view=%+v change=%+v modes=%v err=%v",
			view, fixture.bindings.change, fixture.coordinator.modes, err)
	}
}

func TestScheduleChangeRejectsCrossTenantPostCommandView(t *testing.T) {
	fixture := newApplicationFixture(t)
	replacement := currentSubscription(t, testV7(t))
	fixture.bindings.changeView = &replacement
	_, err := fixture.application.ScheduleChange(context.Background(),
		billingapi.ScheduleChangeCommand{
			Access: fixture.access, IdempotencyKey: "change-cross-tenant",
			SubscriptionID:          fixture.bindings.current.ID.String(),
			TargetOfferingVersionID: testV7(t).String(),
			TargetPriceVersionID:    testV7(t).String(),
			ExpectedVersion:         fixture.bindings.current.View.Version,
		})
	if !errors.Is(err, billingapi.ErrUnavailable) || fixture.resolver.calls != 1 ||
		!commandReadModes(fixture.coordinator.modes) {
		t.Fatalf("cross-tenant post-read error=%v modes=%v resolve=%d",
			err, fixture.coordinator.modes, fixture.resolver.calls)
	}
}

func commandReadModes(modes []platformtx.ScopeMode) bool {
	return len(modes) == 3 && modes[0] == platformtx.SnapshotQuery &&
		modes[1] == platformtx.Command && modes[2] == platformtx.SnapshotQuery
}
