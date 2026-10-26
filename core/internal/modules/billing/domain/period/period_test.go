package period

import (
	"errors"
	"testing"
	"time"
)

func TestPeriodCanonicalizesUTCAndIsHalfOpen(t *testing.T) {
	zone := time.FixedZone("fixture", 3*60*60)
	start := time.Date(2026, time.July, 17, 3, 0, 0, 123456000, zone)
	end := start.Add(time.Hour)
	value, err := New(start, end)
	if err != nil {
		t.Fatal(err)
	}
	wantStart := time.Date(2026, time.July, 17, 0, 0, 0, 123456000, time.UTC)
	if !value.Start().Equal(wantStart) || value.Start().Location() != time.UTC {
		t.Fatalf("canonical start = %s", value.Start())
	}
	if !value.Contains(start) || value.Contains(end) {
		t.Fatal("period is not half-open")
	}
}

func TestPeriodRejectsInvalidPrecisionAndRange(t *testing.T) {
	valid := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		start, end time.Time
		want       error
	}{
		{"sub-microsecond", valid.Add(time.Nanosecond), valid.Add(time.Hour), ErrInstant},
		{"zero", time.Time{}, valid, ErrInstant},
		{"empty", valid, valid, ErrRange},
		{"reversed", valid.Add(time.Hour), valid, ErrRange},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New(test.start, test.end); !errors.Is(err, test.want) {
				t.Fatalf("New error = %v, want %v", err, test.want)
			}
		})
	}
}
