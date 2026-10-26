package recordkernel

import "math/big"

func evaluateIntegerArithmetic(operator string, left, right Value) (Value, error) {
	leftInt, leftErr := valueInteger(left)
	rightInt, rightErr := valueInteger(right)
	if leftErr != nil || rightErr != nil {
		return Value{}, ErrCalculation
	}
	leftBig, rightBig := big.NewInt(leftInt), big.NewInt(rightInt)
	result := new(big.Int)
	switch operator {
	case "add":
		result.Add(leftBig, rightBig)
	case "subtract":
		result.Sub(leftBig, rightBig)
	case "multiply":
		result.Mul(leftBig, rightBig)
	case "divide":
		if rightInt == 0 {
			return Value{}, ErrCalculation
		}
		remainder := new(big.Int)
		result.QuoRem(leftBig, rightBig, remainder)
		if remainder.Sign() != 0 {
			return Value{}, ErrCalculation
		}
	}
	if !result.IsInt64() {
		return Value{}, ErrCalculation
	}
	return buildInteger(result.Int64()), nil
}

func evaluateDecimalArithmetic(
	operator string,
	left, right Value,
	target *DecimalProfile,
) (Value, error) {
	leftRat, leftErr := valueDecimal(left)
	rightRat, rightErr := valueDecimal(right)
	if leftErr != nil || rightErr != nil || target == nil {
		return Value{}, ErrCalculation
	}
	result := new(big.Rat)
	switch operator {
	case "add":
		result.Add(leftRat, rightRat)
	case "subtract":
		result.Sub(leftRat, rightRat)
	case "multiply":
		result.Mul(leftRat, rightRat)
	case "divide":
		if rightRat.Sign() == 0 {
			return Value{}, ErrCalculation
		}
		result.Quo(leftRat, rightRat)
	default:
		return Value{}, ErrCalculation
	}
	if result.Num().BitLen()+result.Denom().BitLen() > MaximumFieldBytes*8 {
		return Value{}, ErrBudgetExceeded
	}
	return buildRoundedDecimal(result, *target)
}
