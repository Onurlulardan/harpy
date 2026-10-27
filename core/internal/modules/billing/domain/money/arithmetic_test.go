package money

import (
	"errors"
	"testing"
)

func TestAddAndSubtractUseCheckedMinorUnits(t *testing.T) {
	left := mustMoney(t, 4_500_000_000_000_000, "USD")
	right := mustMoney(t, 4_500_000_000_000_000, "USD")
	total, err := Add(left, right)
	if err != nil {
		t.Fatal(err)
	}
	if total.MinorUnits() != MaximumAbsoluteMinorUnits {
		t.Fatalf("sum = %d", total.MinorUnits())
	}
	if _, err := Add(total, mustMoney(t, 1, "USD")); !errors.Is(err, ErrAmountOutOfRange) {
		t.Fatalf("overflow error = %v", err)
	}
	if _, err := Subtract(
		mustMoney(t, -MaximumAbsoluteMinorUnits, "USD"), mustMoney(t, 1, "USD"),
	); !errors.Is(err, ErrAmountOutOfRange) {
		t.Fatalf("negative overflow error = %v", err)
	}
	negative, err := Subtract(mustMoney(t, 5, "USD"), mustMoney(t, 8, "USD"))
	if err != nil || negative.MinorUnits() != -3 {
		t.Fatalf("difference = %d, error = %v", negative.MinorUnits(), err)
	}
}

func TestArithmeticRejectsCurrencyMismatchAndInvalidZeroValue(t *testing.T) {
	if _, err := Add(mustMoney(t, 1, "USD"), mustMoney(t, 1, "EUR")); !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("currency mismatch error = %v", err)
	}
	if _, err := Add(Money{}, mustMoney(t, 1, "USD")); err == nil {
		t.Fatal("invalid zero-value Money was accepted")
	}
}

func TestGrossRejectsNegativeAndOverflow(t *testing.T) {
	value, err := Gross(mustMoney(t, 125, "USD"), 4)
	if err != nil || value.MinorUnits() != 500 {
		t.Fatalf("gross = %d, error = %v", value.MinorUnits(), err)
	}
	if _, err := Gross(mustMoney(t, 1, "USD"), -1); !errors.Is(err, ErrNegativeQuantity) {
		t.Fatalf("negative quantity error = %v", err)
	}
	if _, err := Gross(mustMoney(t, -1, "USD"), 1); !errors.Is(err, ErrNegativeGross) {
		t.Fatalf("negative gross error = %v", err)
	}
	if _, err := Gross(mustMoney(t, MaximumAbsoluteMinorUnits, "USD"), 2); !errors.Is(err, ErrAmountOutOfRange) {
		t.Fatalf("gross overflow error = %v", err)
	}
}
