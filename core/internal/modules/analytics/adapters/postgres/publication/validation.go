package publication

import (
	"github.com/google/uuid"
	"regexp"
	"time"
)

var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}

func validMutation(operationID, actorID uuid.UUID, key string, occurredAt time.Time) bool {
	return validUUIDv7(operationID) && actorID != uuid.Nil && tokenPattern.MatchString(key) &&
		!occurredAt.IsZero() && occurredAt.Location() == time.UTC && occurredAt.Nanosecond()%1_000 == 0
}

func validTransition(inputID, actorID uuid.UUID, key string, digest [32]byte,
	revision uint64, occurredAt time.Time,
) bool {
	return validMutation(inputID, actorID, key, occurredAt) && digest != [32]byte{} && revision > 0
}
