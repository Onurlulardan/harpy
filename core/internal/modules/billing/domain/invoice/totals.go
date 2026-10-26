package invoice

import "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"

type Totals struct {
	invoiceTotal  money.Money
	settled       money.Money
	amountDue     money.Money
	creditBalance money.Money
}

func CalculateTotals(lines []Line, settled money.Money) (Totals, error) {
	if len(lines) == 0 || settled.MinorUnits() < 0 {
		return Totals{}, ErrTotals
	}
	zero, err := money.New(0, settled.Currency(), settled.CurrencyRegistryVersion())
	if err != nil {
		return Totals{}, ErrTotals
	}
	total := zero
	seen := make(map[LineIdentifier]struct{}, len(lines))
	for _, line := range lines {
		if line.id == (LineIdentifier{}) {
			return Totals{}, ErrTotals
		}
		if _, duplicate := seen[line.id]; duplicate {
			return Totals{}, ErrTotals
		}
		seen[line.id] = struct{}{}
		total, err = money.Add(total, line.total)
		if err != nil {
			return Totals{}, ErrTotals
		}
	}
	if _, err := money.Add(zero, settled); err != nil {
		return Totals{}, ErrTotals
	}
	amountDue := zero
	creditBalance := zero
	if total.MinorUnits() > settled.MinorUnits() {
		amountDue, err = money.Subtract(total, settled)
	} else {
		creditBalance, err = money.Subtract(settled, total)
	}
	if err != nil {
		return Totals{}, ErrTotals
	}
	return Totals{total, settled, amountDue, creditBalance}, nil
}

func (totals Totals) InvoiceTotal() money.Money  { return totals.invoiceTotal }
func (totals Totals) Settled() money.Money       { return totals.settled }
func (totals Totals) AmountDue() money.Money     { return totals.amountDue }
func (totals Totals) CreditBalance() money.Money { return totals.creditBalance }
