package money

import (
	"errors"
	"testing"
)

func TestNewPinsRegistryExponentAndBounds(t *testing.T) {
	if PolicyName != "fc_money_v1" || RegistryVersion != 1 {
		t.Fatalf("policy identity = %s/v%d", PolicyName, RegistryVersion)
	}
	for _, test := range []struct {
		name     string
		minor    int64
		currency string
		exponent uint8
	}{
		{name: "positive maximum", minor: MaximumAbsoluteMinorUnits, currency: "USD", exponent: 2},
		{name: "negative maximum", minor: -MaximumAbsoluteMinorUnits, currency: "JPY", exponent: 0},
		{name: "four decimals", minor: 1, currency: "CLF", exponent: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := New(test.minor, test.currency, RegistryVersion)
			if err != nil {
				t.Fatal(err)
			}
			if value.MinorUnits() != test.minor || value.Currency() != test.currency ||
				value.CurrencyRegistryVersion() != 1 || value.Exponent() != test.exponent {
				t.Fatalf("money = %+v", value)
			}
		})
	}
}

func TestMoneyOperationsRejectForgedExponent(t *testing.T) {
	forged := mustMoney(t, 1, "USD")
	forged.exponent = 3
	if _, err := Add(forged, mustMoney(t, 1, "USD")); !errors.Is(err, ErrInvalidMoney) {
		t.Fatalf("forged exponent error = %v", err)
	}
}

func TestNewFailsClosedForInvalidMoney(t *testing.T) {
	tests := []struct {
		name     string
		minor    int64
		currency string
		version  int64
		want     error
	}{
		{name: "positive overflow", minor: MaximumAbsoluteMinorUnits + 1, currency: "USD", version: 1, want: ErrAmountOutOfRange},
		{name: "negative overflow", minor: -MaximumAbsoluteMinorUnits - 1, currency: "USD", version: 1, want: ErrAmountOutOfRange},
		{name: "unknown version", currency: "USD", version: 2, want: ErrUnsupportedRegistryVersion},
		{name: "lowercase", currency: "usd", version: 1, want: ErrUnknownCurrency},
		{name: "unknown code", currency: "ZZZ", version: 1, want: ErrUnknownCurrency},
		{name: "N.A. code", currency: "XAU", version: 1, want: ErrMinorUnitNotApplicable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.minor, test.currency, test.version)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestNewRejectsEveryNotApplicableCurrency(t *testing.T) {
	codes := []string{
		"XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XDR",
		"XPD", "XPT", "XSU", "XTS", "XUA", "XXX",
	}
	for _, code := range codes {
		if _, err := New(0, code, RegistryVersion); !errors.Is(err, ErrMinorUnitNotApplicable) {
			t.Fatalf("currency %s error = %v", code, err)
		}
	}
}

func mustMoney(t *testing.T, minor int64, currency string) Money {
	t.Helper()
	value, err := New(minor, currency, RegistryVersion)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
