package subscription

import (
	"time"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

type CancellationSchedule struct {
	effectiveAt time.Time
}

func ScheduleCancellation(
	state State,
	grantPeriod billingperiod.Period,
	now time.Time,
) (CancellationSchedule, error) {
	if state != Active && state != Trialing || !grantPeriod.Contains(now) {
		return CancellationSchedule{}, ErrSchedule
	}
	return CancellationSchedule{effectiveAt: grantPeriod.End()}, nil
}

func (schedule CancellationSchedule) EffectiveAt() time.Time {
	return schedule.effectiveAt
}

func (schedule CancellationSchedule) CanResume(state State, now time.Time) bool {
	return (state == Active || state == Trialing) &&
		canonicalScheduleInstant(now) && now.Before(schedule.effectiveAt)
}

func canonicalScheduleInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}
