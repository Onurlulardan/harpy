package entitlement

import (
	"time"

	"github.com/google/uuid"
)

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func (snapshot Snapshot) valid() bool {
	if snapshot.Definitions == nil || snapshot.Base == nil ||
		!validV7(snapshot.BaseAssignmentID) || !validV7(snapshot.BaseOfferingVersionID) {
		return false
	}
	if snapshot.Purchased == nil {
		return true
	}
	layer := snapshot.Purchased
	return validV7(layer.SubscriptionID) && validV7(layer.OfferingVersionID) &&
		layer.State.Valid() && layer.Values != nil && layer.Overrides != nil
}
