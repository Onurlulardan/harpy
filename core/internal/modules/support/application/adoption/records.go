package adoption

import (
	"time"

	"github.com/google/uuid"
)

func validPage(manifest, expected [32]byte, after, expectedAfter uuid.UUID, count int,
	done bool,
) bool {
	return manifest == expected && after == expectedAfter && count <= PageLimit &&
		(count > 0 || done) && (done || count == PageLimit)
}

func orderedSourceID(previous, current uuid.UUID) bool {
	return current != uuid.Nil && (previous == uuid.Nil || current.String() > previous.String())
}

func storedTime(value time.Time) (time.Time, bool) {
	if value.IsZero() || value.Nanosecond()%1000 != 0 {
		return time.Time{}, false
	}
	return value.UTC(), true
}

func storedOptionalTime(value *time.Time) (*time.Time, bool) {
	if value == nil {
		return nil, true
	}
	normalized, valid := storedTime(*value)
	if !valid {
		return nil, false
	}
	return &normalized, true
}
