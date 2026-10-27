package usage

import (
	"math"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/google/uuid"
)

func validV7(identifier uuid.UUID) bool {
	return identifier != uuid.Nil && identifier.Version() == uuid.Version(7)
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func (snapshot ReservationSnapshot) validFor(request ReserveRequest) bool {
	if snapshot.Definition.Key() != request.EntitlementKey ||
		snapshot.PeriodID != request.PeriodID || !validV7(snapshot.PeriodID) ||
		snapshot.Period.Start().Location() != time.UTC ||
		snapshot.Period.End().Location() != time.UTC ||
		snapshot.Committed < 0 || snapshot.Reserved < 0 || snapshot.Gauge < 0 ||
		snapshot.Revision <= 0 || !snapshot.Definition.Accepts(snapshot.Limit) {
		return false
	}
	switch snapshot.Definition.MeterKind() {
	case entitlement.Counter:
		return snapshot.Gauge == 0 &&
			snapshot.Committed <= math.MaxInt64-snapshot.Reserved
	case entitlement.Gauge:
		return snapshot.Committed == 0 &&
			snapshot.Gauge <= math.MaxInt64-snapshot.Reserved
	default:
		return false
	}
}

func validReserveInput(request ReserveRequest, now time.Time) bool {
	periodValid := validV7(request.PeriodID) && !request.ResolveCurrentPeriod
	if request.ResolveCurrentPeriod {
		periodValid = request.PeriodID == uuid.Nil
	}
	reservationValid := request.ReservationID == uuid.Nil || validV7(request.ReservationID)
	return request.EntitlementKey.Valid() && periodValid && reservationValid &&
		validV7(request.SourceOperationID) && request.Amount > 0 &&
		validInstant(request.ExpiresAt) && request.ExpiresAt.After(now)
}

func (snapshot ReservationSnapshot) current() int64 {
	if snapshot.Definition.MeterKind() == entitlement.Counter {
		return snapshot.Committed + snapshot.Reserved
	}
	return snapshot.Gauge + snapshot.Reserved
}
