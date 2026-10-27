package period

import "time"

type ManualAnchor struct {
	month       time.Month
	day         int
	hour        int
	minute      int
	second      int
	microsecond int
}

func AnchorFrom(origin time.Time) (ManualAnchor, error) {
	canonical, err := canonicalInstant(origin)
	if err != nil {
		return ManualAnchor{}, err
	}
	return ManualAnchor{
		month: canonical.Month(), day: canonical.Day(), hour: canonical.Hour(),
		minute: canonical.Minute(), second: canonical.Second(),
		microsecond: canonical.Nanosecond() / int(time.Microsecond),
	}, nil
}

func BuildManual(start time.Time, anchor ManualAnchor, cadence Cadence) (Period, error) {
	canonicalStart, err := canonicalInstant(start)
	if err != nil {
		return Period{}, err
	}
	if !cadence.valid() {
		return Period{}, ErrCadence
	}
	if !anchor.valid() || !canonicalStart.Equal(anchor.resolveStart(
		canonicalStart.Year(), canonicalStart.Month(), cadence,
	)) {
		return Period{}, ErrAnchor
	}
	var end time.Time
	if cadence == Monthly {
		year, month := nextMonth(canonicalStart.Year(), canonicalStart.Month())
		end = anchor.resolve(year, month)
	} else {
		end = anchor.resolve(canonicalStart.Year()+1, anchor.month)
	}
	return New(canonicalStart, end)
}

func (anchor ManualAnchor) valid() bool {
	return anchor.month >= time.January && anchor.month <= time.December &&
		anchor.day >= 1 && anchor.day <= 31 && anchor.hour >= 0 && anchor.hour <= 23 &&
		anchor.minute >= 0 && anchor.minute <= 59 && anchor.second >= 0 &&
		anchor.second <= 59 && anchor.microsecond >= 0 && anchor.microsecond <= 999999
}

func (anchor ManualAnchor) resolveStart(
	year int,
	month time.Month,
	cadence Cadence,
) time.Time {
	if cadence == Monthly {
		return anchor.resolve(year, month)
	}
	return anchor.resolve(year, anchor.month)
}

func (anchor ManualAnchor) resolve(year int, month time.Month) time.Time {
	day := min(anchor.day, lastDay(year, month))
	return time.Date(
		year, month, day, anchor.hour, anchor.minute, anchor.second,
		anchor.microsecond*int(time.Microsecond), time.UTC,
	)
}

func nextMonth(year int, month time.Month) (int, time.Month) {
	if month == time.December {
		return year + 1, time.January
	}
	return year, month + 1
}

func lastDay(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
