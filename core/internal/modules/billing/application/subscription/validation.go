package subscriptionapp

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

func validID(identifier uuid.UUID) bool { return identifier != uuid.Nil }

func validV7(identifier uuid.UUID) bool {
	return identifier != uuid.Nil && identifier.Version() == uuid.Version(7)
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}

func validDigest(value [32]byte) bool { return value != [32]byte{} }

func canonicalReason(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	return trimmed, trimmed != "" && len(trimmed) <= 512
}

func canonicalOptionalReason(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	return trimmed, len(trimmed) <= 512
}

func copiedInstant(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
