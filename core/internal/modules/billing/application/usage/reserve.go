package usage

import (
	"context"
	"time"

	"github.com/google/uuid"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func (commands *Commands) Reserve(
	ctx context.Context,
	request ReserveRequest,
) (ReserveResult, error) {
	now := commands.clock()
	if !validInstant(now) || !validReserveInput(request, now) {
		return ReserveResult{}, ErrRequest
	}
	if request.ResolveCurrentPeriod {
		periodID, err := commands.ResolvePeriod(ctx, request.EntitlementKey)
		if err != nil {
			return ReserveResult{}, err
		}
		request.PeriodID = periodID
		request.ResolveCurrentPeriod = false
	}
	result, err := commands.store.Reserve(
		ctx, ReserveExecution{Request: request, Now: now},
		func(snapshot ReservationSnapshot, existing *ReservationReplay) (ReservePlan, error) {
			return commands.planReserve(request, now, snapshot, existing)
		},
	)
	if err != nil {
		return ReserveResult{}, usageOperationError("reserve billing usage", err)
	}
	return result, nil
}

func (commands *Commands) planReserve(
	request ReserveRequest,
	now time.Time,
	snapshot ReservationSnapshot,
	existing *ReservationReplay,
) (ReservePlan, error) {
	if !snapshot.validFor(request) {
		return ReservePlan{}, ErrSnapshot
	}
	if existing != nil {
		result, err := replayResult(*existing, request, snapshot)
		return ReservePlan{Result: result}, err
	}
	admission, err := usagedomain.AdmitReservation(
		snapshot.Definition, snapshot.Limit,
		snapshot.Committed, snapshot.Reserved, snapshot.Gauge, request.Amount,
	)
	if err != nil {
		return ReservePlan{}, ErrSnapshot
	}
	decisionID, err := commands.nextID()
	if err != nil || !validV7(decisionID) {
		return ReservePlan{}, ErrConfiguration
	}
	decision := DecisionRecord{
		ID: decisionID, EntitlementKey: request.EntitlementKey,
		PeriodID: request.PeriodID, SourceOperationID: request.SourceOperationID,
		Current: snapshot.current(), Requested: request.Amount,
		Limit: snapshot.Limit, Allowed: admission.Allowed(),
		SoftExceeded: admission.SoftLimitExceeded(), DecidedAt: now,
	}
	if !admission.Allowed() {
		return ReservePlan{Result: decisionResult(decision, snapshot), Decision: &decision}, nil
	}
	reservation, err := commands.newAllowedReservation(request, decision.ID)
	if err != nil {
		return ReservePlan{}, err
	}
	result := decisionResult(decision, snapshot)
	result.Reservation = reservation
	plan := ReservePlan{Result: result, Decision: &decision, Reservation: &reservation}
	if ThresholdEventRequired(decision) {
		event, eventErr := commands.thresholdEvent(
			request.PeriodID, snapshot.Revision+1, request.SourceOperationID, now,
		)
		if eventErr != nil {
			return ReservePlan{}, eventErr
		}
		plan.Event = &event
	}
	return plan, nil
}

func (commands *Commands) newAllowedReservation(
	request ReserveRequest,
	decisionID uuid.UUID,
) (Reservation, error) {
	identifier := request.ReservationID
	if identifier == uuid.Nil {
		var err error
		identifier, err = commands.nextID()
		if err != nil || !validV7(identifier) {
			return Reservation{}, ErrConfiguration
		}
	}
	reservation := Reservation{
		ID: identifier, DecisionID: decisionID, EntitlementKey: request.EntitlementKey,
		PeriodID: request.PeriodID, SourceOperationID: request.SourceOperationID,
		Amount: request.Amount, State: usagedomain.Reserved,
		ExpiresAt: request.ExpiresAt, Revision: 1,
	}
	return reservation, nil
}
