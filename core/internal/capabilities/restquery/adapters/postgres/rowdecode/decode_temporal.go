package rowdecode

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

const timestampLayout = "2006-01-02T15:04:05.000000Z"

func decodeTemporal(
	fieldType api.FieldType, kind api.LogicalType, driverValue any,
) (Value, error) {
	value, ok := driverValue.(time.Time)
	if !ok || value.Year() < 1 || value.Year() > 9999 {
		return Value{}, ErrInvalidValue
	}
	switch kind {
	case api.LogicalTypeDate:
		if value.Location() != time.UTC || value.Hour() != 0 || value.Minute() != 0 ||
			value.Second() != 0 || value.Nanosecond() != 0 {
			return Value{}, ErrInvalidValue
		}
		canonical := value.Format(time.DateOnly)
		parsed, err := time.Parse(time.DateOnly, canonical)
		if err != nil || !parsed.Equal(value) {
			return Value{}, ErrInvalidValue
		}
		return scalarValue(fieldType, kind, canonical)
	case api.LogicalTypeTimestamp:
		if value.Nanosecond()%1000 != 0 {
			return Value{}, ErrInvalidValue
		}
		canonical := value.UTC().Format(timestampLayout)
		parsed, err := time.Parse(timestampLayout, canonical)
		if err != nil || !parsed.Equal(value) {
			return Value{}, ErrInvalidValue
		}
		return scalarValue(fieldType, kind, canonical)
	default:
		return Value{}, ErrInvalidFieldType
	}
}
