package recordkernel

import (
	"math"
	"math/big"
)

func evaluateUnary(
	node expression,
	value Value,
	targetDecimal *DecimalProfile,
) (Value, error) {
	switch node.operator {
	case "not":
		boolean, err := valueBoolean(value)
		if err != nil {
			return Value{}, err
		}
		return buildBoolean(!boolean), nil
	case "negate":
		if node.typeID == TypeInteger {
			integer, err := valueInteger(value)
			if err != nil || integer == math.MinInt64 {
				return Value{}, ErrCalculation
			}
			return buildInteger(-integer), nil
		}
		decimal, err := valueDecimal(value)
		if err != nil {
			return Value{}, err
		}
		if targetDecimal == nil {
			return Value{}, ErrCalculation
		}
		return buildRoundedDecimal(new(big.Rat).Neg(decimal), *targetDecimal)
	default:
		return Value{}, ErrCalculation
	}
}
