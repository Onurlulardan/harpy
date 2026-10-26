package subscriptionapp

import (
	"time"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func dueCancellationSnapshot(
	before SubscriptionSnapshot,
	boundary BoundarySnapshot,
	now time.Time,
) (SubscriptionSnapshot, *ChangeResolution, error) {
	if !before.CancelAtPeriodEnd || before.CancelAt == nil ||
		!before.CancelAt.Equal(boundary.DueAt) ||
		subscriptiondomain.ValidateTransition(before.State, subscriptiondomain.Cancelled) != nil {
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	after := before.revised(now)
	after.State = subscriptiondomain.Cancelled
	after.CancelAtPeriodEnd, after.CancelAt = false, nil
	after.TrialEndsAt, after.GraceEndsAt = nil, nil
	return checkedDueSnapshot(after, nil)
}

func dueGraceSnapshot(
	before SubscriptionSnapshot,
	boundary BoundarySnapshot,
	now time.Time,
) (SubscriptionSnapshot, *ChangeResolution, error) {
	if before.State != subscriptiondomain.PastDue || before.GraceEndsAt == nil ||
		!before.GraceEndsAt.Equal(boundary.DueAt) ||
		subscriptiondomain.ValidateTransition(before.State, subscriptiondomain.Expired) != nil {
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	after := before.revised(now)
	after.State, after.GraceEndsAt = subscriptiondomain.Expired, nil
	after.CancelAtPeriodEnd, after.CancelAt = false, nil
	return checkedDueSnapshot(after, nil)
}
