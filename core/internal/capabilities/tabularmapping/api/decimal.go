package tabularmapping

import (
	"math/big"
	"strings"
)

func finiteDecimal(value *big.Rat) (string, error) {
	if value == nil {
		return "", ErrInvalid
	}
	denominator := new(big.Int).Set(value.Denom())
	twos := removeFactor(denominator, 2)
	fives := removeFactor(denominator, 5)
	if denominator.Cmp(big.NewInt(1)) != 0 {
		return "", ErrInvalid
	}
	scale := max(twos, fives)
	if scale > MaximumDecimalScale {
		return "", ErrInvalid
	}
	power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	factor := new(big.Int).Quo(power, value.Denom())
	scaled := new(big.Int).Mul(value.Num(), factor)
	negative := scaled.Sign() < 0
	digits := new(big.Int).Abs(scaled).String()
	if scale > 0 {
		if len(digits) <= scale {
			digits = strings.Repeat("0", scale-len(digits)+1) + digits
		}
		point := len(digits) - scale
		digits = digits[:point] + "." + digits[point:]
		digits = strings.TrimRight(strings.TrimRight(digits, "0"), ".")
	}
	if digits == "" || digits == "-0" {
		digits = "0"
	}
	if negative && digits != "0" {
		digits = "-" + digits
	}
	return digits, nil
}

func removeFactor(value *big.Int, factor int64) int {
	count := 0
	divisor := big.NewInt(factor)
	remainder := new(big.Int)
	for {
		remainder.Mod(value, divisor)
		if remainder.Sign() != 0 {
			return count
		}
		value.Quo(value, divisor)
		count++
	}
}
