package money

import (
	"fmt"
	"math/big"
)

const (
	PolicyName                      = "fc_money_v1"
	RegistryVersion           int64 = currencyRegistryVersion
	MaximumAbsoluteMinorUnits       = int64(9_000_000_000_000_000)
)

// Money is an immutable, registry-pinned minor-unit amount.
type Money struct {
	minorUnits      int64
	currency        string
	registryVersion int64
	exponent        uint8
}

func New(minorUnits int64, currency string, registryVersion int64) (Money, error) {
	if registryVersion != RegistryVersion {
		return Money{}, fmt.Errorf("%w: %d", ErrUnsupportedRegistryVersion,
			registryVersion)
	}
	entry, ok := LookupCurrency(registryVersion, currency)
	if !ok {
		return Money{}, fmt.Errorf("%w: %q", ErrUnknownCurrency, currency)
	}
	if !entry.MinorUnitApplicable {
		return Money{}, fmt.Errorf("%w: %s", ErrMinorUnitNotApplicable, currency)
	}
	if _, err := checkedMinor(new(big.Int).SetInt64(minorUnits)); err != nil {
		return Money{}, err
	}
	return Money{
		minorUnits:      minorUnits,
		currency:        currency,
		registryVersion: registryVersion,
		exponent:        entry.Exponent,
	}, nil
}

func (value Money) MinorUnits() int64 {
	return value.minorUnits
}

func (value Money) Currency() string {
	return value.currency
}

func (value Money) CurrencyRegistryVersion() int64 {
	return value.registryVersion
}

func (value Money) Exponent() uint8 {
	return value.exponent
}

func (value Money) IsZero() bool {
	return value.minorUnits == 0
}

func (value Money) validate() error {
	if value.registryVersion != RegistryVersion {
		return fmt.Errorf("%w: %d", ErrUnsupportedRegistryVersion,
			value.registryVersion)
	}
	entry, ok := LookupCurrency(value.registryVersion, value.currency)
	if !ok || !entry.MinorUnitApplicable || entry.Exponent != value.exponent {
		return ErrInvalidMoney
	}
	_, err := checkedMinor(new(big.Int).SetInt64(value.minorUnits))
	return err
}
