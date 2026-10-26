package subscriptionapp

import (
	"testing"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func TestBasePlanEmitsEntitlementEventOnlyForGrantChange(t *testing.T) {
	before := fixtureSnapshot(t, subscriptiondomain.Active)
	after := before.revised(fixtureNow)
	unchanged := basePlan(fixturePlanInputs(t, before, after))
	if len(unchanged.Outbox) != 1 ||
		unchanged.Outbox[0].EventType != outboxSubscriptionChanged {
		t.Fatalf("unchanged grant outbox = %#v", unchanged.Outbox)
	}

	changed := after
	changed.State = subscriptiondomain.PastDue
	grace := changed.Period.End()
	changed.GraceEndsAt = &grace
	plan := basePlan(fixturePlanInputs(t, before, changed))
	if len(plan.Outbox) != 2 ||
		plan.Outbox[1].EventType != outboxEntitlementsChanged ||
		plan.Outbox[1].ID == plan.Outbox[0].ID {
		t.Fatalf("changed grant outbox = %#v", plan.Outbox)
	}
}

func fixturePlanInputs(
	t *testing.T,
	before SubscriptionSnapshot,
	after SubscriptionSnapshot,
) planInputs {
	t.Helper()
	return planInputs{
		Before: before, After: after,
		Execution: Execution{SourceOperationID: fixtureV7(t)},
		ReceiptID: fixtureV7(t), AuditID: fixtureV7(t),
		EventID: fixtureV7(t), EntitlementEventID: fixtureV7(t),
		ActorID: fixtureV7(t), Now: fixtureNow,
		Kind: "fixture_change", Reason: "fixture",
	}
}
