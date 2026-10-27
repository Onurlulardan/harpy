package recordkernel

import (
	"encoding/json"
	"github.com/google/uuid"
	"slices"
)

func decodeUUID(typeID TypeID, raw []byte) (Value, error) {
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return Value{}, ErrInputInvalid
	}
	value, err := uuid.Parse(text)
	if err != nil || value == uuid.Nil || value.String() != text {
		return Value{}, ErrInputInvalid
	}
	encoded, _ := json.Marshal(text)
	return typedValue(typeID, encoded), nil
}

func decodeEnum(field Field, raw []byte) (Value, error) {
	var text string
	if json.Unmarshal(raw, &text) != nil || !slices.Contains(field.EnumOptions, text) {
		return Value{}, ErrInputInvalid
	}
	encoded, _ := json.Marshal(text)
	return typedValue(field.TypeID, encoded), nil
}
