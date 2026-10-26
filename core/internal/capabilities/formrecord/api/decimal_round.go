package recordkernel

import (
	"math/big"
	"strings"
)

func roundedDecimal(
	value *big.Rat,
	scale int,
	precision int,
) (string, *big.Rat, error) {
	if value == nil || scale < 0 || precision < 1 || scale > precision {
		return "", nil, ErrCalculation
	}
	power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	numerator := new(big.Int).Mul(new(big.Int).Abs(value.Num()), power)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, value.Denom(), remainder)
	twiceRemainder := new(big.Int).Lsh(remainder, 1)
	comparison := twiceRemainder.Cmp(value.Denom())
	if comparison > 0 || comparison == 0 && quotient.Bit(0) == 1 {
		quotient.Add(quotient, big.NewInt(1))
	}
	digits := quotient.String()
	if len(digits) > precision {
		return "", nil, ErrCalculation
	}
	if scale > 0 && len(digits) <= scale {
		digits = strings.Repeat("0", scale-len(digits)+1) + digits
	}
	text := digits
	if scale > 0 {
		position := len(digits) - scale
		text = digits[:position] + "." + digits[position:]
	}
	if value.Sign() < 0 && quotient.Sign() != 0 {
		text = "-" + text
	}
	rounded, valid := new(big.Rat).SetString(text)
	if !valid {
		return "", nil, ErrCalculation
	}
	return text, rounded, nil
}
