package rowdecode

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

// DecodeRow decodes an exact positional pgx row and never returns a partial row.
func DecodeRow(fieldTypes []api.FieldType, driverValues []any) ([]Value, error) {
	if len(fieldTypes) == 0 || len(fieldTypes) > maximumRowColumns ||
		len(fieldTypes) != len(driverValues) {
		return nil, ErrInvalidRow
	}
	result := make([]Value, len(fieldTypes))
	for index := range fieldTypes {
		value, err := Decode(fieldTypes[index], driverValues[index])
		if err != nil {
			return nil, ErrInvalidRow
		}
		result[index] = value
	}
	return result, nil
}
