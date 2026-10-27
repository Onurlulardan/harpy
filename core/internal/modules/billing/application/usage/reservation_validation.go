package usage

import (
	"github.com/google/uuid"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func validStoredReservation(value Reservation) bool {
	if !validV7(value.ID) || !validV7(value.DecisionID) ||
		!value.EntitlementKey.Valid() || !validV7(value.PeriodID) ||
		!validV7(value.SourceOperationID) || value.Amount <= 0 ||
		!value.State.Valid() || !validInstant(value.ExpiresAt) || value.Revision <= 0 {
		return false
	}
	if (value.EffectStartedAt == nil) != (value.EffectReceiptID == nil) {
		return false
	}
	if value.EffectStartedAt != nil &&
		(!validInstant(*value.EffectStartedAt) || !validV7(*value.EffectReceiptID)) {
		return false
	}
	if value.ResolutionReceipt != nil && !validV7(*value.ResolutionReceipt) {
		return false
	}
	switch value.State {
	case usagedomain.Committed:
		return value.CommittedAmount > 0 && value.CommittedAmount <= value.Amount &&
			(value.ResolutionReceipt != nil || value.EffectStartedAt == nil)
	case usagedomain.Released:
		return value.CommittedAmount == 0 &&
			(value.ResolutionReceipt != nil || value.EffectStartedAt == nil)
	case usagedomain.Reserved, usagedomain.ReconcileRequired:
		return value.CommittedAmount == 0
	default:
		return false
	}
}

func sameReceipt(stored, requested *uuid.UUID) bool {
	if stored == nil || requested == nil {
		return stored == nil && requested == nil
	}
	return *stored == *requested
}
