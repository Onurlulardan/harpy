package billing

import (
	"time"

	"github.com/google/uuid"
)

const utcMicrosecondLayout = "2006-01-02T15:04:05.000000Z"

func parseUUID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value
}

func firstError(candidate error, fallback error) error {
	if candidate != nil {
		return candidate
	}
	return fallback
}

func validUTCInterval(start string, end *string) bool {
	startAt, err := time.Parse(utcMicrosecondLayout, start)
	if err != nil || startAt.Format(utcMicrosecondLayout) != start {
		return false
	}
	if end == nil {
		return true
	}
	endAt, err := time.Parse(utcMicrosecondLayout, *end)
	return err == nil && endAt.Format(utcMicrosecondLayout) == *end && endAt.After(startAt)
}

func positiveOptional(value *int64) bool {
	return value == nil || *value > 0
}

func validOfferingEligibility(all bool, identifiers []string) bool {
	if all {
		return len(identifiers) == 0
	}
	return validUUIDv7Set(identifiers, 1, 256)
}

func validUUIDv7Set(identifiers []string, minimum int, maximum int) bool {
	if len(identifiers) < minimum || len(identifiers) > maximum {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		parsed, valid := parseUUID(identifier)
		if !valid || parsed.Version() != 7 {
			return false
		}
		if _, duplicate := seen[parsed]; duplicate {
			return false
		}
		seen[parsed] = struct{}{}
	}
	return true
}

func validCurrency(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, character := range value {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}

func validEntitlementKey(value string) bool {
	if len(value) < 1 || len(value) > 128 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	last := value[len(value)-1]
	if last == '.' || last == '_' {
		return false
	}
	for _, character := range value[1:] {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '_' && character != '.' {
			return false
		}
	}
	return true
}
