package rowdecode

import (
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
	"github.com/google/uuid"
)

func decodeList(
	fieldType api.FieldType, kind api.LogicalType, driverValue any,
) (Value, error) {
	elements, ok := driverValue.([]any)
	if !ok {
		return Value{}, ErrInvalidValue
	}
	values := make([]string, len(elements))
	for index, element := range elements {
		var valid bool
		switch kind {
		case api.LogicalTypeTextList:
			values[index], valid = textElement(element)
		case api.LogicalTypeUUIDList:
			values[index], valid = uuidElement(element)
		}
		if !valid {
			return Value{}, ErrInvalidValue
		}
	}
	response, err := json.Marshal(values)
	if err != nil {
		return Value{}, ErrInvalidValue
	}
	return newValue(fieldType, kind, response, zeroLastValue(), false, false)
}

func textElement(value any) (string, bool) {
	text, ok := value.(string)
	if !ok || textprofile.ExactNFC(text) != text {
		return "", false
	}
	return text, validText(text)
}

func uuidElement(value any) (string, bool) {
	bytes, ok := value.([16]byte)
	if !ok {
		return "", false
	}
	identifier := uuid.UUID(bytes)
	if identifier == uuid.Nil || identifier.Variant() != uuid.RFC4122 {
		return "", false
	}
	return identifier.String(), true
}
