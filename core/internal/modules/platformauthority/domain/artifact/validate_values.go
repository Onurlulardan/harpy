package artifact

import (
	"encoding/base64"
	"regexp"
	"strings"
	"time"
)

var (
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	kidPattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9._:/-]{0,511}$`)
	uuidPattern   = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
)

func validUUID(value string) bool {
	return uuidPattern.MatchString(value) && value != "00000000-0000-0000-0000-000000000000"
}

func validDigest(value string) bool { return digestPattern.MatchString(value) }

func validOptionalDigest(value *string) bool {
	return value == nil || validDigest(*value)
}

func validKID(value string) bool {
	return kidPattern.MatchString(value) && strings.TrimSpace(value) == value
}

func parseInstant(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil && parsed.Location() == time.UTC &&
		parsed.Nanosecond() == 0 && parsed.Format(time.RFC3339) == value
}

func decodeBase64URL(value string, size int) ([]byte, bool) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return decoded, err == nil && len(decoded) == size &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
}
