package invoice

import "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"

type LineIdentifier [16]byte

type Line struct {
	id               LineIdentifier
	gross            money.Money
	discount         money.Money
	tax              money.Money
	debitAdjustment  money.Money
	creditAdjustment money.Money
	total            money.Money
}

func NewLine(
	id LineIdentifier,
	gross money.Money,
	discount money.Money,
	tax money.Money,
	debitAdjustment money.Money,
	creditAdjustment money.Money,
) (Line, error) {
	if id == (LineIdentifier{}) {
		return Line{}, ErrLine
	}
	zero, err := money.New(0, gross.Currency(), gross.CurrencyRegistryVersion())
	if err != nil {
		return Line{}, ErrLine
	}
	components := []money.Money{gross, discount, tax, debitAdjustment, creditAdjustment}
	for _, component := range components {
		if component.MinorUnits() < 0 {
			return Line{}, ErrLine
		}
		if _, err := money.Add(zero, component); err != nil {
			return Line{}, ErrLine
		}
	}
	if discount.MinorUnits() > gross.MinorUnits() {
		return Line{}, ErrLine
	}
	total, err := calculateTotal(gross, discount, tax, debitAdjustment, creditAdjustment)
	if err != nil {
		return Line{}, ErrLine
	}
	return Line{id, gross, discount, tax, debitAdjustment, creditAdjustment, total}, nil
}

func NewVerifiedLine(
	id LineIdentifier,
	gross money.Money,
	discount money.Money,
	tax money.Money,
	debitAdjustment money.Money,
	creditAdjustment money.Money,
	assertedTotal money.Money,
) (Line, error) {
	line, err := NewLine(id, gross, discount, tax, debitAdjustment, creditAdjustment)
	if err != nil {
		return Line{}, err
	}
	if line.total.MinorUnits() != assertedTotal.MinorUnits() ||
		line.total.Currency() != assertedTotal.Currency() ||
		line.total.CurrencyRegistryVersion() != assertedTotal.CurrencyRegistryVersion() {
		return Line{}, ErrInvariant
	}
	return line, nil
}

func calculateTotal(
	gross money.Money,
	discount money.Money,
	tax money.Money,
	debit money.Money,
	credit money.Money,
) (money.Money, error) {
	net, err := money.Subtract(gross, discount)
	if err != nil {
		return money.Money{}, err
	}
	net, err = money.Add(net, tax)
	if err != nil {
		return money.Money{}, err
	}
	net, err = money.Add(net, debit)
	if err != nil {
		return money.Money{}, err
	}
	return money.Subtract(net, credit)
}

func (line Line) Identifier() LineIdentifier    { return line.id }
func (line Line) Gross() money.Money            { return line.gross }
func (line Line) Discount() money.Money         { return line.discount }
func (line Line) Tax() money.Money              { return line.tax }
func (line Line) DebitAdjustment() money.Money  { return line.debitAdjustment }
func (line Line) CreditAdjustment() money.Money { return line.creditAdjustment }
func (line Line) Total() money.Money            { return line.total }
