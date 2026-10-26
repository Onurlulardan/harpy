package offering

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

func validActor(value uuid.UUID) bool { return value != uuid.Nil }

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func validIdempotencyKey(value string) bool {
	return value != "" && len(value) <= 128 && strings.TrimSpace(value) == value
}

func validReason(value string) bool {
	return value != "" && len(value) <= 512 && strings.TrimSpace(value) == value
}

func validSnapshot(value Snapshot, expectVersion bool) bool {
	if !validV7(value.OfferingID) || value.Revision < 1 ||
		(value.Lifecycle != "active" && value.Lifecycle != "retired") {
		return false
	}
	if expectVersion {
		return validV7(value.VersionID) && value.Version > 0 && value.Lifecycle == "active"
	}
	return value.VersionID == uuid.Nil && value.Version == 0 && value.Lifecycle == "retired"
}

func nextV7(nextID IDGenerator) (uuid.UUID, error) {
	value, err := nextID()
	if err != nil || !validV7(value) {
		return uuid.Nil, ErrConfiguration
	}
	return value, nil
}

func newPriceIDs(nextID IDGenerator, count int) ([]uuid.UUID, error) {
	identifiers := make([]uuid.UUID, count)
	for index := range identifiers {
		identifier, err := nextV7(nextID)
		if err != nil {
			return nil, err
		}
		identifiers[index] = identifier
	}
	return identifiers, nil
}
