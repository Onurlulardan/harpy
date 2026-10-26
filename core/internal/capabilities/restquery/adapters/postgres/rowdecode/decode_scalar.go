package rowdecode

import (
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func decodeText(
	fieldType api.FieldType, kind api.LogicalType, driverValue any,
) (Value, error) {
	text, ok := driverValue.(string)
	if !ok {
		return Value{}, ErrInvalidValue
	}
	return scalarValue(fieldType, kind, text)
}

func decodeBoolean(fieldType api.FieldType, driverValue any) (Value, error) {
	boolean, ok := driverValue.(bool)
	if !ok {
		return Value{}, ErrInvalidValue
	}
	resolvedValue, err := resolved.NewBooleanValue(boolean)
	if err != nil {
		return Value{}, ErrInvalidValue
	}
	last, err := cursor.NewLastValue(resolvedValue)
	if err != nil {
		return Value{}, ErrInvalidValue
	}
	response, err := json.Marshal(boolean)
	if err != nil {
		return Value{}, ErrInvalidValue
	}
	return newValue(fieldType, api.LogicalTypeBoolean, response, last, true, false)
}

func scalarValue(
	fieldType api.FieldType, kind api.LogicalType, canonical string,
) (Value, error) {
	resolvedValue, err := resolved.NewScalarValue(kind, canonical)
	if err != nil {
		return Value{}, ErrInvalidValue
	}
	last, err := cursor.NewLastValue(resolvedValue)
	if err != nil {
		return Value{}, ErrInvalidValue
	}
	response, err := json.Marshal(canonical)
	if err != nil {
		return Value{}, ErrInvalidValue
	}
	return newValue(fieldType, kind, response, last, true, false)
}
