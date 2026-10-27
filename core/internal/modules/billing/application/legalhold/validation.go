package legalhold

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}

func validText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}

func validSnapshot(value Snapshot, tenantID uuid.UUID) bool {
	if !validV7(value.HoldID) || value.TenantID != tenantID || value.Revision < 1 ||
		!validV7(value.PlaceOperationID) || value.PlacedBy == uuid.Nil ||
		!validText(value.Reason, 2048) || !validText(value.Ticket, 256) ||
		!validInstant(value.PlacedAt) {
		return false
	}
	if value.State == "active" {
		return value.Revision == 1 && value.ReleasedBy == uuid.Nil &&
			value.ReleaseOperationID == uuid.Nil && value.ReleasedAt.IsZero() &&
			value.ReleaseReason == ""
	}
	return value.State == "released" && value.Revision == 2 &&
		value.ReleasedBy != uuid.Nil && validV7(value.ReleaseOperationID) &&
		validText(value.ReleaseReason, 2048) && validInstant(value.ReleasedAt) &&
		!value.ReleasedAt.Before(value.PlacedAt)
}

func nextV7(source IDSource) (uuid.UUID, error) {
	value, err := source()
	if err != nil || !validV7(value) {
		return uuid.Nil, ErrConfiguration
	}
	return value, nil
}
