package subscriptionapp

import "time"

func dueSnapshot(
	before SubscriptionSnapshot,
	boundary BoundarySnapshot,
	now time.Time,
) (SubscriptionSnapshot, *ChangeResolution, error) {
	switch boundary.Kind {
	case BoundaryCancellation:
		return dueCancellationSnapshot(before, boundary, now)
	case BoundaryManualChange:
		return dueChangeSnapshot(before, boundary, now)
	case BoundaryTrialEnd:
		return dueTrialSnapshot(before, boundary, now)
	case BoundaryGraceEnd:
		return dueGraceSnapshot(before, boundary, now)
	case BoundaryPeriodEnd:
		return duePeriodSnapshot(before, boundary, now)
	default:
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
}

func checkedDueSnapshot(
	after SubscriptionSnapshot,
	resolution *ChangeResolution,
) (SubscriptionSnapshot, *ChangeResolution, error) {
	if !after.valid() || resolution != nil && !resolution.valid() {
		return SubscriptionSnapshot{}, nil, ErrSnapshot
	}
	return after, resolution, nil
}
