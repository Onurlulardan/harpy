package campaign

import "time"

func canonicalInstant(value time.Time) (time.Time, error) {
	if value.IsZero() || value.Location() != time.UTC ||
		value.Nanosecond()%int(time.Microsecond) != 0 {
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

func contains(start time.Time, end *time.Time, instant time.Time) bool {
	return !instant.Before(start) && (end == nil || instant.Before(*end))
}
