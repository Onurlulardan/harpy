package money

import (
	"errors"
	"testing"
)

func TestApplyBasisPointsRoundsHalfToEven(t *testing.T) {
	rate, err := NewBasisPoints(5_000)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		minor int64
		want  int64
	}{
		{minor: 1, want: 0},
		{minor: 3, want: 2},
		{minor: 5, want: 2},
		{minor: 7, want: 4},
		{minor: -1, want: 0},
		{minor: -3, want: -2},
	}
	for _, test := range tests {
		result, err := ApplyBasisPoints(mustMoney(t, test.minor, "USD"), rate)
		if err != nil || result.MinorUnits() != test.want {
			t.Fatalf("half-even(%d) = %d, error = %v", test.minor,
				result.MinorUnits(), err)
		}
	}
}

func TestBasisPointsValidateEveryUse(t *testing.T) {
	for _, value := range []int64{-1, MaximumBasisPoints + 1} {
		if _, err := NewBasisPoints(value); !errors.Is(err, ErrInvalidBasisPoints) {
			t.Fatalf("basis points %d error = %v", value, err)
		}
	}
	if _, err := ApplyBasisPoints(
		mustMoney(t, 100, "USD"), BasisPoints(MaximumBasisPoints+1),
	); !errors.Is(err, ErrInvalidBasisPoints) {
		t.Fatalf("cast basis points error = %v", err)
	}
}
