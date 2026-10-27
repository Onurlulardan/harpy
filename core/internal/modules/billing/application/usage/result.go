package usage

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
	"github.com/google/uuid"
)

func decisionResult(decision DecisionRecord, snapshot ReservationSnapshot) ReserveResult {
	return ReserveResult{
		Allowed: decision.Allowed, SoftLimitExceeded: decision.SoftExceeded,
		Current: decision.Current, Requested: decision.Requested,
		Limit: decision.Limit, Period: snapshot.Period,
	}
}

func replayResult(
	replay ReservationReplay,
	request ReserveRequest,
	snapshot ReservationSnapshot,
) (ReserveResult, error) {
	reservation := replay.Reservation
	decision := replay.Decision
	if decision.EntitlementKey != request.EntitlementKey ||
		decision.PeriodID != request.PeriodID ||
		decision.SourceOperationID != request.SourceOperationID ||
		decision.Requested != request.Amount ||
		!sameValue(decision.Limit, snapshot.Limit) {
		return ReserveResult{}, ErrReservationReused
	}
	if !decision.Allowed {
		if reservation != (Reservation{}) {
			return ReserveResult{}, ErrReservationReused
		}
		result := decisionResult(decision, snapshot)
		result.Replay = true
		return result, nil
	}
	if reservation.EntitlementKey != request.EntitlementKey ||
		request.ReservationID != uuid.Nil && reservation.ID != request.ReservationID ||
		reservation.PeriodID != request.PeriodID ||
		reservation.SourceOperationID != request.SourceOperationID ||
		reservation.Amount != request.Amount ||
		!reservation.ExpiresAt.Equal(request.ExpiresAt) ||
		!validStoredReservation(reservation) || reservation.DecisionID != decision.ID ||
		!decision.Allowed {
		return ReserveResult{}, ErrReservationReused
	}
	if reservation.State != usagedomain.Reserved &&
		reservation.State != usagedomain.Committed &&
		reservation.State != usagedomain.Released &&
		reservation.State != usagedomain.ReconcileRequired {
		return ReserveResult{}, ErrReservationReused
	}
	result := decisionResult(decision, snapshot)
	result.Reservation = reservation
	result.Replay = true
	return result, nil
}

func sameValue(left, right entitlement.Value) bool {
	if left.Kind() != right.Kind() || left.Unlimited() != right.Unlimited() {
		return false
	}
	if leftBoolean, ok := left.Boolean(); ok {
		rightBoolean, rightOK := right.Boolean()
		return rightOK && leftBoolean == rightBoolean
	}
	leftInteger, leftOK := left.Integer()
	rightInteger, rightOK := right.Integer()
	return leftOK == rightOK && (!leftOK || leftInteger == rightInteger)
}
