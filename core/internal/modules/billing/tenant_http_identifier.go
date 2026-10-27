package billing

import (
	"github.com/google/uuid"
)

func parseV7(value string) (uuid.UUID, bool) {
	identifier, err := uuid.Parse(value)
	return identifier, err == nil && identifier.Version() == uuid.Version(7)
}
