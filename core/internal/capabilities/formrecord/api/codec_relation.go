package recordkernel

import (
	"bytes"
	"encoding/json"
)

func decodeRelation(field Field, raw []byte) (Value, int, error) {
	if field.Relation == nil {
		return Value{}, 0, ErrInputInvalid
	}
	if bytes.Equal(raw, []byte("null")) {
		value, err := DecodeValue(field, raw)
		return value, 0, err
	}
	if field.Relation.Cardinality == "one" {
		value, err := DecodeValue(field, raw)
		return value, 1, err
	}
	var elements []json.RawMessage
	if json.Unmarshal(raw, &elements) != nil || elements == nil ||
		len(elements) > MaximumRelationEdges {
		return Value{}, 0, ErrBudgetExceeded
	}
	seen := make(map[string]struct{}, len(elements))
	canonical := make([]json.RawMessage, len(elements))
	for index, element := range elements {
		elementField := field
		elementField.TypeID = TypeUUID
		elementField.Nullable = false
		value, err := DecodeValue(elementField, element)
		if err != nil {
			return Value{}, 0, err
		}
		key := string(value.canonical)
		if _, duplicate := seen[key]; duplicate {
			return Value{}, 0, ErrInputInvalid
		}
		seen[key] = struct{}{}
		canonical[index] = value.CanonicalJSON()
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return Value{}, 0, ErrInputInvalid
	}
	return typedValue(TypeRelation, encoded), len(elements), nil
}
