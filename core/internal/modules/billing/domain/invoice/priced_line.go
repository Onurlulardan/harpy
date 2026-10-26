package invoice

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

type PricedLineInput struct {
	Identifier       LineIdentifier
	UnitPrice        money.Money
	Quantity         int64
	DiscountRules    []discount.Rule
	CycleOrdinal     int64
	TaxRate          money.BasisPoints
	DebitAdjustment  money.Money
	CreditAdjustment money.Money
}

type PricedLine struct {
	line         Line
	applications []discount.Application
}

func PriceLine(input PricedLineInput) (PricedLine, error) {
	gross, err := money.Gross(input.UnitPrice, input.Quantity)
	if err != nil {
		return PricedLine{}, ErrLine
	}
	discountResult, err := discount.Apply(gross, input.CycleOrdinal, input.DiscountRules)
	if err != nil {
		return PricedLine{}, ErrLine
	}
	tax, err := money.ApplyBasisPoints(discountResult.Net(), input.TaxRate)
	if err != nil {
		return PricedLine{}, ErrLine
	}
	line, err := NewLine(
		input.Identifier, gross, discountResult.Discount(), tax,
		input.DebitAdjustment, input.CreditAdjustment,
	)
	if err != nil {
		return PricedLine{}, err
	}
	return PricedLine{line: line, applications: discountResult.Applications()}, nil
}

func (priced PricedLine) Line() Line { return priced.line }

func (priced PricedLine) DiscountApplications() []discount.Application {
	return append([]discount.Application(nil), priced.applications...)
}
