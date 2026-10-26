package rowdecode

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func decodeUUID(
	fieldType api.FieldType, kind api.LogicalType, driverValue any,
) (Value, error) {
	bytes, ok := driverValue.([16]byte)
	if !ok {
		return Value{}, ErrInvalidValue
	}
	identifier := uuid.UUID(bytes)
	if identifier == uuid.Nil || identifier.Variant() != uuid.RFC4122 {
		return Value{}, ErrInvalidValue
	}
	return scalarValue(fieldType, kind, identifier.String())
}

func decodeEnum(fieldType api.FieldType, driverValue any) (Value, error) {
	code, ok := driverValue.(string)
	if !ok {
		return Value{}, ErrInvalidValue
	}
	profile, found, err := fieldType.Enum()
	if err != nil || !found {
		return Value{}, ErrInvalidFieldType
	}
	_, declared, err := profile.Lookup(code)
	if err != nil {
		return Value{}, ErrInvalidFieldType
	}
	if !declared {
		return Value{}, ErrInvalidValue
	}
	return scalarValue(fieldType, api.LogicalTypeEnum, code)
}
