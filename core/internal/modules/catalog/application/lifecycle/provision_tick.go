package lifecycle

import "time"

const applyTickBudget = 15 * time.Second

func provisionTickDecision(
	ready bool,
	retryAt time.Time,
	started time.Time,
	now time.Time,
	remaining bool,
) (time.Time, bool) {
	wait := time.Time{}
	if !ready {
		wait = retryAt
	}
	stop := remaining && !now.Before(started.Add(applyTickBudget))
	return wait, stop
}

func earliestProvisionWait(current, next time.Time) time.Time {
	if next.IsZero() {
		return current
	}
	if current.IsZero() || next.Before(current) {
		return next
	}
	return current
}

func floorProvisionRetry(retryAt, now time.Time) time.Time {
	if retryAt.IsZero() {
		return time.Time{}
	}
	minimum := now.Add(time.Second).UTC().Truncate(time.Microsecond)
	if !retryAt.After(now) {
		return minimum
	}
	return retryAt.UTC().Truncate(time.Microsecond)
}
