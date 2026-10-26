package subscriptionapp

import (
	"time"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func dueTrialSnapshot(
	before SubscriptionSnapshot,
	boundary BoundarySnapshot,
	now time.Time,
) (SubscriptionSnapshot, *ChangeResolution, error) {
	if before.State != subscriptiondomain.Trialing || before.TrialEndsAt == nil ||
		!before.TrialEndsAt.Equal(boundary.DueAt) {
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	after := before.revised(now)
	after.TrialEndsAt, after.CancelAt = nil, nil
	after.CancelAtPeriodEnd = false
	switch {
	case boundary.PaymentReceiptID != nil && boundary.NextPeriod != nil &&
		boundary.NextPeriod.Start().Equal(boundary.DueAt) && boundary.NextPeriod.End().After(now):
		after.State, after.Period = subscriptiondomain.Active, boundary.NextPeriod
		after.GraceEndsAt = nil
	case boundary.PaymentReceiptID == nil && boundary.GraceEndsAt != nil &&
		boundary.GraceEndsAt.After(now):
		after.State = subscriptiondomain.PastDue
		after.GraceEndsAt = copiedInstant(boundary.GraceEndsAt)
	case boundary.PaymentReceiptID == nil:
		after.State, after.GraceEndsAt = subscriptiondomain.Expired, nil
	default:
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	if subscriptiondomain.ValidateTransition(before.State, after.State) != nil {
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	return checkedDueSnapshot(after, nil)
}
