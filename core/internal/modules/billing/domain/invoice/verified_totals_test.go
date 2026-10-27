package invoice

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func TestNewVerifiedTotalsRejectsArithmeticMismatch(t *testing.T) {
	amount := func(units int64) money.Money {
		value, err := money.New(units, "USD", money.RegistryVersion)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	result, err := NewVerifiedTotals(amount(1000), amount(400), amount(600), amount(0))
	if err != nil || result.AmountDue().MinorUnits() != 600 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, err := NewVerifiedTotals(amount(1000), amount(400), amount(599), amount(0)); err == nil {
		t.Fatal("mismatched amount due was accepted")
	}
}
