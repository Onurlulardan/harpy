package invoice

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func TestPriceLineUsesGrossDiscountTaxAdjustmentOrder(t *testing.T) {
	rate, _ := money.NewBasisPoints(1000)
	taxRate, _ := money.NewBasisPoints(2000)
	rule, err := discount.NewBasisPointRule(
		discountID(1), 1, rate, discount.ForeverDuration())
	if err != nil {
		t.Fatal(err)
	}
	priced, err := PriceLine(PricedLineInput{
		Identifier: lineID(1), UnitPrice: amount(t, 500, "USD"), Quantity: 2,
		DiscountRules: []discount.Rule{rule}, CycleOrdinal: 1, TaxRate: taxRate,
		DebitAdjustment:  amount(t, 30, "USD"),
		CreditAdjustment: amount(t, 10, "USD"),
	})
	if err != nil {
		t.Fatal(err)
	}
	line := priced.Line()
	if line.Gross().MinorUnits() != 1000 || line.Discount().MinorUnits() != 100 ||
		line.Tax().MinorUnits() != 180 || line.Total().MinorUnits() != 1100 ||
		len(priced.DiscountApplications()) != 1 {
		t.Fatalf("unexpected priced line: %+v", line)
	}
}

func TestPriceLineRejectsTaxOnPreDiscountAmount(t *testing.T) {
	rate, _ := money.NewBasisPoints(5000)
	taxRate, _ := money.NewBasisPoints(1000)
	rule, _ := discount.NewBasisPointRule(
		discountID(1), 1, rate, discount.ForeverDuration())
	priced, err := PriceLine(PricedLineInput{
		Identifier: lineID(1), UnitPrice: amount(t, 101, "USD"), Quantity: 1,
		DiscountRules: []discount.Rule{rule}, CycleOrdinal: 1, TaxRate: taxRate,
		DebitAdjustment:  amount(t, 0, "USD"),
		CreditAdjustment: amount(t, 0, "USD"),
	})
	if err != nil || priced.Line().Discount().MinorUnits() != 50 ||
		priced.Line().Tax().MinorUnits() != 5 || priced.Line().Total().MinorUnits() != 56 {
		t.Fatalf("line=%+v err=%v", priced.Line(), err)
	}
}

func discountID(last byte) discount.Identifier {
	var identifier discount.Identifier
	identifier[len(identifier)-1] = last
	return identifier
}
