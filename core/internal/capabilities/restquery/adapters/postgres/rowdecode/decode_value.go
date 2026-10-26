package rowdecode

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
)

// Decode converts an exact pgx driver value without text or float coercion.
func Decode(fieldType api.FieldType, driverValue any) (Value, error) {
	if err := fieldType.Validate(); err != nil {
		return Value{}, ErrInvalidFieldType
	}
	kind, err := fieldType.Kind()
	if err != nil {
		return Value{}, ErrInvalidFieldType
	}
	if driverValue == nil {
		return decodeNull(fieldType, kind)
	}
	switch kind {
	case api.LogicalTypeText:
		return decodeText(fieldType, kind, driverValue)
	case api.LogicalTypeInteger:
		return decodeInteger(fieldType, driverValue)
	case api.LogicalTypeDecimal:
		return decodeDecimal(fieldType, driverValue)
	case api.LogicalTypeBoolean:
		return decodeBoolean(fieldType, driverValue)
	case api.LogicalTypeDate, api.LogicalTypeTimestamp:
		return decodeTemporal(fieldType, kind, driverValue)
	case api.LogicalTypeUUID, api.LogicalTypeRelation:
		return decodeUUID(fieldType, kind, driverValue)
	case api.LogicalTypeEnum:
		return decodeEnum(fieldType, driverValue)
	case api.LogicalTypeTextList, api.LogicalTypeUUIDList:
		return decodeList(fieldType, kind, driverValue)
	case api.LogicalTypeJSON:
		return decodeJSON(fieldType, driverValue)
	default:
		return Value{}, ErrInvalidFieldType
	}
}

func decodeNull(fieldType api.FieldType, kind api.LogicalType) (Value, error) {
	last := cursor.LastValue{}
	hasLast := cursorKind(kind)
	if hasLast {
		last = cursor.NewNullLastValue()
	}
	return newValue(fieldType, kind, []byte("null"), last, hasLast, true)
}

func cursorKind(kind api.LogicalType) bool {
	return kind >= api.LogicalTypeText && kind <= api.LogicalTypeRelation
}
