package domain

import "time"

type QuietHours struct {
	TimeZone               string
	StartMinute, EndMinute int
}

func (value QuietHours) Valid() bool {
	if value.TimeZone == "" || value.StartMinute < 0 || value.StartMinute >= 24*60 ||
		value.EndMinute < 0 || value.EndMinute >= 24*60 ||
		value.StartMinute == value.EndMinute {
		return false
	}
	_, err := time.LoadLocation(value.TimeZone)
	return err == nil
}

func (value QuietHours) MaterializeNotBefore(now time.Time) (time.Time, error) {
	if !value.Valid() || now.IsZero() {
		return time.Time{}, ErrPreferenceInvalid
	}
	location, _ := time.LoadLocation(value.TimeZone)
	local := now.In(location)
	minute := local.Hour()*60 + local.Minute()
	inside := value.StartMinute < value.EndMinute &&
		minute >= value.StartMinute && minute < value.EndMinute ||
		value.StartMinute > value.EndMinute &&
			(minute >= value.StartMinute || minute < value.EndMinute)
	if !inside {
		return now.UTC(), nil
	}
	day := local
	if value.StartMinute > value.EndMinute && minute >= value.StartMinute {
		day = local.AddDate(0, 0, 1)
	}
	hour, minuteOfHour := value.EndMinute/60, value.EndMinute%60
	end := time.Date(day.Year(), day.Month(), day.Day(), hour, minuteOfHour, 0, 0, location)
	return end.UTC(), nil
}
