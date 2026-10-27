package invoice

import "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"

func NewVerifiedTotals(
	invoiceTotal money.Money,
	settled money.Money,
	amountDue money.Money,
	creditBalance money.Money,
) (Totals, error) {
	zero, err := money.New(
		0, invoiceTotal.Currency(), invoiceTotal.CurrencyRegistryVersion(),
	)
	if err != nil || settled.MinorUnits() < 0 || amountDue.MinorUnits() < 0 ||
		creditBalance.MinorUnits() < 0 {
		return Totals{}, ErrTotals
	}
	for _, value := range []money.Money{settled, amountDue, creditBalance} {
		if _, err := money.Add(zero, value); err != nil {
			return Totals{}, ErrTotals
		}
	}
	expectedDue, expectedCredit := zero, zero
	if invoiceTotal.MinorUnits() > settled.MinorUnits() {
		expectedDue, err = money.Subtract(invoiceTotal, settled)
	} else {
		expectedCredit, err = money.Subtract(settled, invoiceTotal)
	}
	if err != nil || expectedDue.MinorUnits() != amountDue.MinorUnits() ||
		expectedCredit.MinorUnits() != creditBalance.MinorUnits() {
		return Totals{}, ErrTotals
	}
	return Totals{invoiceTotal, settled, amountDue, creditBalance}, nil
}
