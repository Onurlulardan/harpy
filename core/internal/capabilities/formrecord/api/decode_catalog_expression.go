package recordkernel

import (
	"encoding/json"
	"github.com/google/uuid"
)

func decodeCatalogExpression(
	document catalogExpression,
	expected TypeID,
	fields map[uuid.UUID]Field,
	limits BudgetLimits,
) (expression, uint64, map[uuid.UUID]struct{}, error) {
	if !validCatalogExpressionShape(document) {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	typed := typedExpressionDocument(document)
	return compileTypedExpression(typed, expected, fields, 1, limits)
}

func validCatalogExpressionShape(value catalogExpression) bool {
	leafBase := value.FieldID == "" && value.Operator == "" && value.Function == ""
	switch value.Kind {
	case "literal":
		return leafBase && len(value.Value) > 0 && value.Arguments == nil
	case "field":
		return value.FieldID != "" && value.Operator == "" && value.Function == "" &&
			value.Value == nil && value.Arguments == nil
	case "unary", "binary":
		return value.FieldID == "" && value.Operator != "" && value.Function == "" &&
			value.Value == nil && value.Arguments != nil
	case "call":
		return value.FieldID == "" && value.Operator == "" && value.Function != "" &&
			value.Value == nil && value.Arguments != nil
	default:
		return false
	}
}

func typedExpressionDocument(value catalogExpression) TypedExpression {
	result := TypedExpression{Kind: value.Kind, Type: value.Type}
	if value.FieldID != "" {
		copy := value.FieldID
		result.FieldID = &copy
	}
	if len(value.Value) > 0 {
		result.Value = append(json.RawMessage(nil), value.Value...)
	}
	if value.Operator != "" {
		copy := value.Operator
		result.Operator = &copy
	}
	if value.Function != "" {
		copy := value.Function
		result.Function = &copy
	}
	if value.Arguments != nil {
		result.Arguments = make([]TypedExpression, len(value.Arguments))
		for index, argument := range value.Arguments {
			result.Arguments[index] = typedExpressionDocument(argument)
		}
	}
	return result
}
