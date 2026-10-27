package recordkernel

import (
	"github.com/google/uuid"
)

func legacyChildTypes(
	source Expression,
	expected TypeID,
	fields map[uuid.UUID]Field,
) ([]TypeID, error) {
	if source.Kind == "unary" && source.Operator != nil {
		if *source.Operator == "not" && expected == TypeBoolean {
			return []TypeID{TypeBoolean}, nil
		}
		if *source.Operator == "negate" && numericType(expected) {
			return []TypeID{expected}, nil
		}
	}
	if source.Kind == "binary" && source.Operator != nil {
		return legacyBinaryTypes(source, expected, fields)
	}
	if source.Kind == "call" && source.Function != nil {
		return legacyCallTypes(source, expected)
	}
	return nil, ErrCatalogInvalid
}

func legacyBinaryTypes(
	source Expression,
	expected TypeID,
	fields map[uuid.UUID]Field,
) ([]TypeID, error) {
	if len(source.Arguments) != 2 {
		return nil, ErrCatalogInvalid
	}
	switch *source.Operator {
	case "add", "subtract", "multiply", "divide":
		if numericType(expected) {
			return []TypeID{expected, expected}, nil
		}
	case "and", "or":
		if expected == TypeBoolean {
			return []TypeID{expected, expected}, nil
		}
	case "eq", "ne", "gt", "gte", "lt", "lte":
		if expected != TypeBoolean {
			break
		}
		hint, ok := legacyComparisonType(source.Arguments, fields)
		if ok {
			return []TypeID{hint, hint}, nil
		}
	}
	return nil, ErrCatalogInvalid
}

func legacyComparisonType(
	arguments []Expression,
	fields map[uuid.UUID]Field,
) (TypeID, bool) {
	var result TypeID
	for _, argument := range arguments {
		hint, found := legacyFieldHint(argument, fields)
		if !found {
			continue
		}
		if result != "" && result != hint {
			return "", false
		}
		result = hint
	}
	_, supported := expressionType(string(result))
	return result, result != "" && supported
}

func legacyFieldHint(value Expression, fields map[uuid.UUID]Field) (TypeID, bool) {
	if value.Kind == "field" && value.FieldID != nil {
		fieldID, err := parseStableID(*value.FieldID)
		field, found := fields[fieldID]
		return field.TypeID, err == nil && found
	}
	for _, argument := range value.Arguments {
		if hint, found := legacyFieldHint(argument, fields); found {
			return hint, true
		}
	}
	return "", false
}
