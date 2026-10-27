package billing

import "time"

const canonicalTimestamp = "2006-01-02T15:04:05.000000Z"

func parseTimestamp(value string) (time.Time, error) {
	instant, err := time.Parse(canonicalTimestamp, value)
	if err != nil || instant.Format(canonicalTimestamp) != value {
		return time.Time{}, errInvalidInput
	}
	return instant.UTC(), nil
}

func parseOptionalTimestamp(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	instant, err := parseTimestamp(*value)
	if err != nil {
		return nil, err
	}
	return &instant, nil
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(canonicalTimestamp)
}

func formatOptionalTimestamp(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatTimestamp(*value)
	return &formatted
}
