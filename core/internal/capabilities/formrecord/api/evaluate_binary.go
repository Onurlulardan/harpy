package recordkernel

func evaluateBinary(
	node expression,
	left, right Value,
	targetDecimal *DecimalProfile,
) (Value, error) {
	switch node.operator {
	case "add", "subtract", "multiply", "divide":
		if node.typeID == TypeInteger {
			return evaluateIntegerArithmetic(node.operator, left, right)
		}
		return evaluateDecimalArithmetic(node.operator, left, right, targetDecimal)
	case "eq", "ne", "gt", "gte", "lt", "lte":
		return evaluateComparison(node.operator, left, right)
	case "and", "or":
		return evaluateBooleanBinary(node.operator, left, right)
	default:
		return Value{}, ErrCalculation
	}
}

func evaluateBooleanBinary(operator string, left, right Value) (Value, error) {
	leftValue, leftErr := valueBoolean(left)
	rightValue, rightErr := valueBoolean(right)
	if leftErr != nil || rightErr != nil {
		return Value{}, ErrCalculation
	}
	if operator == "and" {
		return buildBoolean(leftValue && rightValue), nil
	}
	return buildBoolean(leftValue || rightValue), nil
}
