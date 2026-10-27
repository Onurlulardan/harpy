package domain

import (
	"testing"
	"time"
)

func TestCalendarWindowSeparatesTimedAndAllDayModels(t *testing.T) {
	start := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	if err := (CalendarWindow{StartsAt: &start, EndsAt: &end, TimeZone: "Europe/Istanbul"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (CalendarWindow{StartDate: "2026-03-01", EndDate: "2026-03-02", TimeZone: "Europe/Istanbul"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (CalendarWindow{StartsAt: &start, EndsAt: &end,
		StartDate: "2026-03-01", EndDate: "2026-03-02", TimeZone: "UTC"}).Validate(); err == nil {
		t.Fatal("mixed time model was accepted")
	}
}
