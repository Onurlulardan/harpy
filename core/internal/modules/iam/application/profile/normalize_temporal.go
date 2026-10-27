package profile

import "time"

const profileTimestampLayout = "2006-01-02T15:04:05.000000Z"

func normalizeTemporalValue(kind LogicalType, document []byte) ([]byte, error) {
	value, err := decodeString(document)
	if err != nil {
		return nil, ErrValidationInvalid
	}
	valid := false
	if kind == LogicalDate {
		parsed, parseErr := time.Parse(time.DateOnly, value)
		valid = parseErr == nil && parsed.Format(time.DateOnly) == value
	} else if kind == LogicalTimestamp {
		parsed, parseErr := time.Parse(profileTimestampLayout, value)
		valid = parseErr == nil && parsed.UTC().Format(profileTimestampLayout) == value
	}
	if !valid {
		return nil, ErrValidationInvalid
	}
	return encodeScalar(value)
}
