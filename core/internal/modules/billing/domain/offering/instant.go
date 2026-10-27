package offering

import "time"

func canonicalInstant(value time.Time) (time.Time, error) {
	if value.IsZero() || value.Nanosecond()%int(time.Microsecond) != 0 {
		return time.Time{}, ErrInstant
	}
	return time.UnixMicro(value.UnixMicro()).UTC(), nil
}

func canonicalEnd(value *time.Time, start time.Time) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	end, err := canonicalInstant(*value)
	if err != nil || !start.Before(end) {
		return nil, ErrInstant
	}
	return &end, nil
}
