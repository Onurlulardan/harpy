package dashboard

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)
var namePattern = regexp.MustCompile(`^[\pL\pN][\pL\pN _.-]{0,99}$`)
var surfacePattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,63}$`)

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}

func parseUUIDv7(value string) (uuid.UUID, error) {
	result, err := uuid.Parse(value)
	if err != nil || !validUUIDv7(result) {
		return uuid.Nil, errInvalidUUID{}
	}
	return result, nil
}

type errInvalidUUID struct{}

func (errInvalidUUID) Error() string { return "invalid UUIDv7" }

func validMutation(operationID, actorID uuid.UUID, key string, occurredAt time.Time) bool {
	return validUUIDv7(operationID) && actorID != uuid.Nil && tokenPattern.MatchString(key) &&
		!occurredAt.IsZero() && occurredAt.Location() == time.UTC &&
		occurredAt.Nanosecond()%1_000 == 0
}
