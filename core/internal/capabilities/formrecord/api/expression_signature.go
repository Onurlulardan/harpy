package recordkernel

func validExpressionSignature(value expression) bool {
	types := argumentTypes(value.arguments)
	switch value.kind {
	case "literal", "field":
		return len(types) == 0
	case "unary":
		return validUnary(value, types)
	case "binary":
		return validBinary(value, types)
	case "call":
		return validCall(value, types)
	default:
		return false
	}
}

func validUnary(value expression, types []TypeID) bool {
	if len(types) != 1 {
		return false
	}
	return value.operator == "not" && value.typeID == TypeBoolean &&
		types[0] == TypeBoolean || value.operator == "negate" &&
		numericType(value.typeID) && types[0] == value.typeID
}

func validBinary(value expression, types []TypeID) bool {
	if len(types) != 2 {
		return false
	}
	same := types[0] == types[1]
	switch value.operator {
	case "add", "subtract", "multiply", "divide":
		return numericType(value.typeID) && same && types[0] == value.typeID
	case "eq", "ne":
		return value.typeID == TypeBoolean && same
	case "gt", "gte", "lt", "lte":
		return value.typeID == TypeBoolean && same && orderedType(types[0])
	case "and", "or":
		return value.typeID == TypeBoolean && same &&
			types[0] == TypeBoolean
	default:
		return false
	}
}

func validCall(value expression, types []TypeID) bool {
	switch value.function {
	case "coalesce":
		return len(types) >= 1 && len(types) <= MaximumExpressionArguments &&
			allType(types, value.typeID)
	case "concat":
		return value.typeID == TypeText && len(types) >= 1 &&
			allType(types, TypeText)
	case "lower", "upper":
		return value.typeID == TypeText && len(types) == 1 &&
			types[0] == TypeText
	case "round":
		return value.typeID == TypeDecimal && len(types) == 2 &&
			types[0] == TypeDecimal && types[1] == TypeInteger
	case "date_add":
		return (value.typeID == TypeDate ||
			value.typeID == TypeTimestamp) && len(types) == 2 &&
			types[0] == value.typeID && types[1] == TypeInteger
	default:
		return false
	}
}

func argumentTypes(values []expression) []TypeID {
	result := make([]TypeID, len(values))
	for index, value := range values {
		result[index] = value.typeID
	}
	return result
}

func allType(values []TypeID, expected TypeID) bool {
	for _, value := range values {
		if value != expected {
			return false
		}
	}
	return true
}
