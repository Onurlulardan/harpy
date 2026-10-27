package subscriptionapp

import (
	"time"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func nextBoundary(snapshot SubscriptionSnapshot) (BoundaryKind, time.Time, bool) {
	if snapshot.CancelAtPeriodEnd && snapshot.CancelAt != nil {
		return BoundaryCancellation, *snapshot.CancelAt, true
	}
	switch snapshot.State {
	case subscriptiondomain.Trialing:
		if snapshot.TrialEndsAt != nil {
			return BoundaryTrialEnd, *snapshot.TrialEndsAt, true
		}
	case subscriptiondomain.Active:
		if snapshot.Period != nil {
			return BoundaryPeriodEnd, snapshot.Period.End(), true
		}
	case subscriptiondomain.PastDue:
		if snapshot.GraceEndsAt != nil {
			return BoundaryGraceEnd, *snapshot.GraceEndsAt, true
		}
	}
	return "", time.Time{}, false
}
