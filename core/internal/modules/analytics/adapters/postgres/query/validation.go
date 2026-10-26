package query

import (
	"regexp"
	"time"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)
var namePattern = regexp.MustCompile(`^[\pL\pN][\pL\pN _.-]{0,99}$`)

func validMutation(operationID, actorID uuid.UUID, key string, occurredAt time.Time) bool {
	return validUUIDv7(operationID) && actorID != uuid.Nil &&
		idempotencyPattern.MatchString(key) && !occurredAt.IsZero() &&
		occurredAt.Location() == time.UTC && occurredAt.Nanosecond()%1_000 == 0
}

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122 && value.Version() == 7
}

func validDocuments(values ...queryapp.Document) bool {
	for _, value := range values {
		if !value.Valid() {
			return false
		}
	}
	return true
}
