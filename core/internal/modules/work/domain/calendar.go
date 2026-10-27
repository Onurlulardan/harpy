package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrCalendarWindow = errors.New("work calendar window is invalid")

type Audience string

const (
	AudiencePrivate      Audience = "PRIVATE"
	AudienceBoard        Audience = "BOARD"
	AudienceOrganization Audience = "ORGANIZATION"
)

func (audience Audience) Valid() bool {
	return audience == AudiencePrivate || audience == AudienceBoard ||
		audience == AudienceOrganization
}

type CalendarWindow struct {
	StartsAt, EndsAt   *time.Time
	StartDate, EndDate string
	TimeZone           string
}

func (window CalendarWindow) Validate() error {
	timed := window.StartsAt != nil || window.EndsAt != nil
	allDay := window.StartDate != "" || window.EndDate != ""
	if timed == allDay || strings.TrimSpace(window.TimeZone) == "" {
		return ErrCalendarWindow
	}
	if timed {
		if window.StartsAt == nil || window.EndsAt == nil ||
			!window.EndsAt.After(*window.StartsAt) ||
			window.StartsAt.Location() != time.UTC || window.EndsAt.Location() != time.UTC {
			return ErrCalendarWindow
		}
		return nil
	}
	start, startErr := time.Parse("2006-01-02", window.StartDate)
	end, endErr := time.Parse("2006-01-02", window.EndDate)
	if startErr != nil || endErr != nil || !end.After(start) {
		return ErrCalendarWindow
	}
	return nil
}
