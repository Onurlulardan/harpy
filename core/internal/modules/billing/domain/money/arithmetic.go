package money

import "fmt"

func Add(left, right Money) (Money, error) {
	if err := requireSameCurrency(left, right); err != nil {
		return Money{}, err
	}
	minorUnits, err := checkedAdd(left.minorUnits, right.minorUnits)
	if err != nil {
		return Money{}, err
	}
	return New(minorUnits, left.currency, left.registryVersion)
}

func Subtract(left, right Money) (Money, error) {
	if err := requireSameCurrency(left, right); err != nil {
		return Money{}, err
	}
	minorUnits, err := checkedSubtract(left.minorUnits, right.minorUnits)
	if err != nil {
		return Money{}, err
	}
	return New(minorUnits, left.currency, left.registryVersion)
}

// Gross multiplies a non-negative unit price by a non-negative quantity.
func Gross(unitPrice Money, quantity int64) (Money, error) {
	if err := unitPrice.validate(); err != nil {
		return Money{}, err
	}
	if quantity < 0 {
		return Money{}, ErrNegativeQuantity
	}
	if unitPrice.minorUnits < 0 {
		return Money{}, ErrNegativeGross
	}
	minorUnits, err := checkedMultiply(unitPrice.minorUnits, quantity)
	if err != nil {
		return Money{}, err
	}
	return New(minorUnits, unitPrice.currency, unitPrice.registryVersion)
}

func requireSameCurrency(left, right Money) error {
	if err := left.validate(); err != nil {
		return err
	}
	if err := right.validate(); err != nil {
		return err
	}
	if left.currency != right.currency ||
		left.registryVersion != right.registryVersion || left.exponent != right.exponent {
		return fmt.Errorf("%w: %s/v%d and %s/v%d", ErrCurrencyMismatch,
			left.currency, left.registryVersion, right.currency, right.registryVersion)
	}
	return nil
}
