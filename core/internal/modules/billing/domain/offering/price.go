package offering

import "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"

type Cadence string

const (
	OneTime Cadence = "one_time"
	Monthly Cadence = "monthly"
	Yearly  Cadence = "yearly"
)

type Price struct {
	cadence Cadence
	amount  money.Money
}

func NewPrice(cadence Cadence, amount money.Money) (Price, error) {
	if !validCadence(cadence) || amount.MinorUnits() < 0 {
		return Price{}, ErrPrice
	}
	verified, err := money.New(
		amount.MinorUnits(), amount.Currency(), amount.CurrencyRegistryVersion(),
	)
	if err != nil || verified.Exponent() != amount.Exponent() {
		return Price{}, ErrPrice
	}
	return Price{cadence: cadence, amount: amount}, nil
}

func (price Price) Cadence() Cadence    { return price.cadence }
func (price Price) Amount() money.Money { return price.amount }

func validCadence(cadence Cadence) bool {
	return cadence == OneTime || cadence == Monthly || cadence == Yearly
}
