package cursor

import "github.com/google/uuid"

func parseCanonicalUUID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value &&
		parsed.Variant() == uuid.RFC4122
}
