package recordkernel

func evaluateCall(
	node expression,
	values []Value,
	targetDecimal *DecimalProfile,
) (Value, error) {
	switch node.function {
	case "concat":
		return evaluateConcat(values)
	case "lower":
		return evaluateTextCase(values[0], false)
	case "upper":
		return evaluateTextCase(values[0], true)
	case "round":
		return evaluateRound(values[0], values[1], targetDecimal)
	case "date_add":
		return evaluateDateAdd(node.typeID, values[0], values[1])
	default:
		return Value{}, ErrCalculation
	}
}
