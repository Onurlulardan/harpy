package subscription

import (
	"testing"
	"time"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

func TestCancellationScheduleUsesExactGrantBoundary(t *testing.T) {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	period, _ := billingperiod.New(start, end)
	schedule, err := ScheduleCancellation(Active, period, start)
	if err != nil || !schedule.EffectiveAt().Equal(end) ||
		!schedule.CanResume(Active, end.Add(-time.Microsecond)) ||
		schedule.CanResume(Active, end) {
		t.Fatalf("schedule=%+v err=%v", schedule, err)
	}
}

func TestCancellationScheduleRejectsNonGrantStates(t *testing.T) {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	period, _ := billingperiod.New(start, start.Add(time.Hour))
	for _, state := range []State{Pending, PastDue, Paused, Cancelled, Expired} {
		if _, err := ScheduleCancellation(state, period, start); err != ErrSchedule {
			t.Fatalf("state %s: expected ErrSchedule, got %v", state, err)
		}
	}
}
