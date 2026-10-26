package subscriptionapp

import (
	"time"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func duePeriodSnapshot(
	before SubscriptionSnapshot,
	boundary BoundarySnapshot,
	now time.Time,
) (SubscriptionSnapshot, *ChangeResolution, error) {
	if before.State != subscriptiondomain.Active || before.Period == nil ||
		!before.Period.End().Equal(boundary.DueAt) || before.CancelAtPeriodEnd {
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	after := before.revised(now)
	if boundary.PaymentReceiptID != nil && boundary.NextPeriod != nil &&
		boundary.NextPeriod.Start().Equal(boundary.DueAt) && boundary.NextPeriod.End().After(now) {
		after.Period = boundary.NextPeriod
	} else if boundary.PaymentReceiptID == nil && boundary.NextPeriod == nil {
		after.State = subscriptiondomain.Expired
		if subscriptiondomain.ValidateTransition(before.State, after.State) != nil {
			return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
		}
	} else {
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	return checkedDueSnapshot(after, nil)
}
