package recordkernel

func evaluateComparison(operator string, left, right Value) (Value, error) {
	comparison, err := compareValues(left, right)
	if err != nil {
		return Value{}, err
	}
	var result bool
	switch operator {
	case "eq":
		result = comparison == 0
	case "ne":
		result = comparison != 0
	case "gt":
		result = comparison > 0
	case "gte":
		result = comparison >= 0
	case "lt":
		result = comparison < 0
	case "lte":
		result = comparison <= 0
	default:
		return Value{}, ErrCalculation
	}
	return buildBoolean(result), nil
}

func compareValues(left, right Value) (int, error) {
	if left.typeID != right.typeID {
		return 0, ErrCalculation
	}
	switch left.typeID {
	case TypeInteger:
		leftValue, leftErr := valueInteger(left)
		rightValue, rightErr := valueInteger(right)
		if leftErr != nil || rightErr != nil {
			return 0, ErrCalculation
		}
		if leftValue < rightValue {
			return -1, nil
		}
		if leftValue > rightValue {
			return 1, nil
		}
		return 0, nil
	case TypeDecimal:
		leftValue, leftErr := valueDecimal(left)
		rightValue, rightErr := valueDecimal(right)
		if leftErr != nil || rightErr != nil {
			return 0, ErrCalculation
		}
		return leftValue.Cmp(rightValue), nil
	case TypeBoolean:
		leftValue, leftErr := valueBoolean(left)
		rightValue, rightErr := valueBoolean(right)
		if leftErr != nil || rightErr != nil {
			return 0, ErrCalculation
		}
		if leftValue == rightValue {
			return 0, nil
		}
		if !leftValue {
			return -1, nil
		}
		return 1, nil
	default:
		leftValue, leftErr := valueString(left)
		rightValue, rightErr := valueString(right)
		if leftErr != nil || rightErr != nil {
			return 0, ErrCalculation
		}
		return compareText(leftValue, rightValue), nil
	}
}
