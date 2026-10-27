package billingv1

import (
	"testing"
	"time"
)

func TestV1JSONInstantMatchesTimeMarshalJSONGoldenCases(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		nano int
		want string
	}{
		{"whole-second", 0, "2026-06-01T02:03:04Z"},
		{"millisecond", 123_000_000, "2026-06-01T02:03:04.123Z"},
		{"trimmed-microsecond", 123_400_000, "2026-06-01T02:03:04.1234Z"},
		{"microsecond", 123_456_000, "2026-06-01T02:03:04.123456Z"},
		{"nanosecond", 123_456_789, "2026-06-01T02:03:04.123456789Z"},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := time.Date(2026, 6, 1, 2, 3, 4, test.nano, time.UTC)
			if got := v1JSONInstant(value); got != test.want {
				t.Fatalf("V1 instant = %q, want %q", got, test.want)
			}
		})
	}
}

func TestV1WireInstantConvertsCanonicalPrecision(t *testing.T) {
	t.Parallel()
	got, err := v1WireInstant("2026-06-01T02:03:04.123400Z")
	if err != nil || got != "2026-06-01T02:03:04.1234Z" {
		t.Fatalf("V1 wire instant = %q, %v", got, err)
	}
	for _, invalid := range []string{
		"2026-06-01T02:03:04Z",
		"2026-06-01T02:03:04.123400+03:00",
		"not-an-instant",
	} {
		if _, err := v1WireInstant(invalid); err == nil {
			t.Fatalf("invalid canonical instant %q was accepted", invalid)
		}
	}
}
