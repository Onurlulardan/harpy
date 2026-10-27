package discount

import "testing"

func TestDurationCycleEligibility(t *testing.T) {
	repeating, err := RepeatingDuration(3)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		duration Duration
		cycle    int64
		want     bool
	}{
		{OnceDuration(), 1, true},
		{OnceDuration(), 2, false},
		{repeating, 3, true},
		{repeating, 4, false},
		{ForeverDuration(), 100, true},
	}
	for _, test := range tests {
		got, err := test.duration.Applies(test.cycle)
		if err != nil || got != test.want {
			t.Fatalf("%s cycle %d: got %v, %v", test.duration.Kind(), test.cycle, got, err)
		}
	}
}

func TestDurationRejectsInvalidValues(t *testing.T) {
	if _, err := RepeatingDuration(0); err != ErrDuration {
		t.Fatalf("expected ErrDuration, got %v", err)
	}
	if _, err := OnceDuration().Applies(0); err != ErrDuration {
		t.Fatalf("expected ErrDuration, got %v", err)
	}
}
