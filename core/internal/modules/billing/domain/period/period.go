package period

import "time"

type Period struct {
	start time.Time
	end   time.Time
}

func New(start, end time.Time) (Period, error) {
	canonicalStart, err := canonicalInstant(start)
	if err != nil {
		return Period{}, err
	}
	canonicalEnd, err := canonicalInstant(end)
	if err != nil {
		return Period{}, err
	}
	if !canonicalStart.Before(canonicalEnd) {
		return Period{}, ErrRange
	}
	return Period{start: canonicalStart, end: canonicalEnd}, nil
}

func (value Period) Start() time.Time {
	return value.start
}

func (value Period) End() time.Time {
	return value.end
}

func (value Period) Contains(instant time.Time) bool {
	canonical, err := canonicalInstant(instant)
	return err == nil && !canonical.Before(value.start) && canonical.Before(value.end)
}
