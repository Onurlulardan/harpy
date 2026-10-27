package invoice

import "testing"

func TestInvoiceTotalsSettlementInvariant(t *testing.T) {
	first, err := NewLine(
		lineID(1), amount(t, 1000, "USD"), amount(t, 100, "USD"),
		amount(t, 180, "USD"), amount(t, 0, "USD"), amount(t, 0, "USD"),
	)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewLine(
		lineID(2), amount(t, 500, "USD"), amount(t, 0, "USD"),
		amount(t, 100, "USD"), amount(t, 0, "USD"), amount(t, 0, "USD"),
	)
	if err != nil {
		t.Fatal(err)
	}
	totals, err := CalculateTotals([]Line{first, second}, amount(t, 1000, "USD"))
	if err != nil || totals.InvoiceTotal().MinorUnits() != 1680 ||
		totals.AmountDue().MinorUnits() != 680 || !totals.CreditBalance().IsZero() {
		t.Fatalf("got %+v, %v", totals, err)
	}
	credit, err := CalculateTotals([]Line{first}, amount(t, 1200, "USD"))
	if err != nil || credit.AmountDue().MinorUnits() != 0 ||
		credit.CreditBalance().MinorUnits() != 120 {
		t.Fatalf("credit got %+v, %v", credit, err)
	}
}

func TestInvoiceTotalsRejectsDuplicateAndCurrencyMismatch(t *testing.T) {
	line, err := NewLine(
		lineID(1), amount(t, 100, "USD"), amount(t, 0, "USD"),
		amount(t, 0, "USD"), amount(t, 0, "USD"), amount(t, 0, "USD"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CalculateTotals([]Line{line, line}, amount(t, 0, "USD")); err != ErrTotals {
		t.Fatalf("duplicate: expected ErrTotals, got %v", err)
	}
	if _, err := CalculateTotals([]Line{line}, amount(t, 0, "EUR")); err != ErrTotals {
		t.Fatalf("currency: expected ErrTotals, got %v", err)
	}
}
