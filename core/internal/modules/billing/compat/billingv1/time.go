package billingv1

import (
	"fmt"
	"time"
)

const utcMicrosecondLayout = "2006-01-02T15:04:05.000000Z"

func utcMicrosecond(value time.Time) string {
	return value.UTC().Truncate(time.Microsecond).Format(utcMicrosecondLayout)
}

func normalizeOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	normalized := utcMicrosecond(*value)
	return &normalized
}

func canonicalWireInstant(value string) bool {
	_, err := v1WireInstant(value)
	return err == nil
}

func v1JSONInstant(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func v1WireInstant(value string) (string, error) {
	parsed, err := time.Parse(utcMicrosecondLayout, value)
	if err != nil || parsed.UTC().Format(utcMicrosecondLayout) != value {
		return "", fmt.Errorf("canonical UTC microsecond instant is invalid")
	}
	return v1JSONInstant(parsed), nil
}

func optionalV1WireInstant(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized, err := v1WireInstant(*value)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}
