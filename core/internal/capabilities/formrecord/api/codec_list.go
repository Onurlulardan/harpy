package recordkernel

import (
	"encoding/json"
)

func decodeList(field Field, raw []byte) (Value, error) {
	var elements []json.RawMessage
	if json.Unmarshal(raw, &elements) != nil || elements == nil ||
		len(elements) > MaximumListElements {
		return Value{}, ErrInputInvalid
	}
	canonical := make([]json.RawMessage, len(elements))
	for index, element := range elements {
		elementField := field
		switch field.TypeID {
		case TypeTextList:
			elementField.TypeID = TypeText
		case TypeUUIDList:
			elementField.TypeID = TypeUUID
		case TypeDocumentRefList:
			elementField.TypeID = TypeDocumentRef
		case TypeEnumList:
			elementField.TypeID = TypeEnum
		default:
			return Value{}, ErrInputInvalid
		}
		elementField.Nullable = false
		value, err := DecodeValue(elementField, element)
		if err != nil {
			return Value{}, ErrInputInvalid
		}
		canonical[index] = value.CanonicalJSON()
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return Value{}, ErrInputInvalid
	}
	return typedValue(field.TypeID, encoded), nil
}
