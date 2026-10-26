package recordkernel

import ()

func legacyCallTypes(
	source Expression,
	expected TypeID,
) ([]TypeID, error) {
	count := len(source.Arguments)
	switch *source.Function {
	case "coalesce":
		if count >= 1 && count <= MaximumExpressionArguments {
			return repeatedType(expected, count), nil
		}
	case "concat":
		if expected == TypeText && count >= 1 &&
			count <= MaximumExpressionArguments {
			return repeatedType(TypeText, count), nil
		}
	case "lower", "upper":
		if expected == TypeText && count == 1 {
			return []TypeID{TypeText}, nil
		}
	case "round":
		if expected == TypeDecimal && count == 2 {
			return []TypeID{TypeDecimal, TypeInteger}, nil
		}
	case "date_add":
		if (expected == TypeDate || expected == TypeTimestamp) && count == 2 {
			return []TypeID{expected, TypeInteger}, nil
		}
	}
	return nil, ErrCatalogInvalid
}

func repeatedType(typeID TypeID, count int) []TypeID {
	result := make([]TypeID, count)
	for index := range result {
		result[index] = typeID
	}
	return result
}

func validLegacyShape(source Expression) bool {
	switch source.Kind {
	case "unary", "binary":
		return source.Operator != nil && source.FieldID == nil &&
			source.Literal == nil && source.Function == nil
	case "call":
		return source.Function != nil && source.FieldID == nil &&
			source.Literal == nil && source.Operator == nil
	default:
		return false
	}
}
