package money

import (
	"fmt"
	"math/big"
)

func checkedMinor(value *big.Int) (int64, error) {
	magnitude := new(big.Int).Abs(new(big.Int).Set(value))
	maximum := new(big.Int).SetInt64(MaximumAbsoluteMinorUnits)
	if magnitude.Cmp(maximum) > 0 || !value.IsInt64() {
		return 0, fmt.Errorf("%w: %s", ErrAmountOutOfRange, value.String())
	}
	return value.Int64(), nil
}

func checkedAdd(left, right int64) (int64, error) {
	result := new(big.Int).Add(
		new(big.Int).SetInt64(left), new(big.Int).SetInt64(right))
	return checkedMinor(result)
}

func checkedSubtract(left, right int64) (int64, error) {
	result := new(big.Int).Sub(
		new(big.Int).SetInt64(left), new(big.Int).SetInt64(right))
	return checkedMinor(result)
}

func checkedMultiply(left, right int64) (int64, error) {
	result := new(big.Int).Mul(
		new(big.Int).SetInt64(left), new(big.Int).SetInt64(right))
	return checkedMinor(result)
}

func roundedProductRatio(value, multiplier, divisor int64) (int64, error) {
	if divisor <= 0 {
		return 0, ErrInvalidMoney
	}
	numerator := new(big.Int).Mul(
		new(big.Int).SetInt64(value), new(big.Int).SetInt64(multiplier))
	result := roundHalfEven(numerator, new(big.Int).SetInt64(divisor))
	return checkedMinor(result)
}

func roundHalfEven(numerator, denominator *big.Int) *big.Int {
	sign := numerator.Sign()
	absolute := new(big.Int).Abs(new(big.Int).Set(numerator))
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(absolute, denominator, remainder)
	twiceRemainder := new(big.Int).Lsh(new(big.Int).Set(remainder), 1)
	comparison := twiceRemainder.Cmp(denominator)
	if comparison > 0 || comparison == 0 && quotient.Bit(0) == 1 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if sign < 0 {
		quotient.Neg(quotient)
	}
	return quotient
}
