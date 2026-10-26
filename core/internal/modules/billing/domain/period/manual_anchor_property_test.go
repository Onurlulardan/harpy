package period

import (
	"testing"
	"time"
)

func TestManualAnchorCalendarProperties(t *testing.T) {
	for year := 1996; year <= 2036; year++ {
		for month := time.January; month <= time.December; month++ {
			day := lastDay(year, month)
			origin := time.Date(year, month, day, 23, 59, 59, 999999000, time.UTC)
			anchor, err := AnchorFrom(origin)
			if err != nil {
				t.Fatal(err)
			}
			monthly, err := BuildManual(origin, anchor, Monthly)
			if err != nil {
				t.Fatalf("monthly %s: %v", origin, err)
			}
			if !monthly.Start().Before(monthly.End()) ||
				monthly.End().Location() != time.UTC || monthly.End().Nanosecond()%1000 != 0 {
				t.Fatalf("monthly invariant failed for %s: %s", origin, monthly.End())
			}
			if _, err := BuildManual(monthly.End(), anchor, Monthly); err != nil {
				t.Fatalf("monthly continuation %s: %v", monthly.End(), err)
			}
			yearly, err := BuildManual(origin, anchor, Yearly)
			if err != nil {
				t.Fatalf("yearly %s: %v", origin, err)
			}
			if !yearly.Start().Before(yearly.End()) {
				t.Fatalf("yearly invariant failed for %s", origin)
			}
			if _, err := BuildManual(yearly.End(), anchor, Yearly); err != nil {
				t.Fatalf("yearly continuation %s: %v", yearly.End(), err)
			}
		}
	}
}
