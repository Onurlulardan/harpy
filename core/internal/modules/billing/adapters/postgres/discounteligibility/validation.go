package discounteligibility

import (
	"time"

	"github.com/google/uuid"
)

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}
