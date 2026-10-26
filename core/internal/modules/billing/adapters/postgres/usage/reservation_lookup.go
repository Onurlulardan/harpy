package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
	"github.com/google/uuid"
)

func (storage *store) findReservation(
	ctx context.Context,
	key entitlement.Key,
	sourceOperationID uuid.UUID,
) (usageapp.ReservationReplay, bool, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return usageapp.ReservationReplay{}, false, err
	}
	reservation := usageapp.Reservation{}
	decision := usageapp.DecisionRecord{}
	var reservationKey, decisionKey, state string
	var limitAmount *int64
	var unlimited, hasReservation bool
	err = storage.queryRow(ctx, findReservationSQL, authorityID, storage.tenantID,
		string(key), sourceOperationID,
	).Scan(&hasReservation, &reservation.ID, &reservation.DecisionID, &reservationKey,
		&reservation.PeriodID, &reservation.SourceOperationID, &reservation.Amount,
		&state, &reservation.ExpiresAt, &reservation.EffectStartedAt,
		&reservation.EffectReceiptID, &reservation.ResolutionReceipt,
		&reservation.CommittedAmount, &reservation.Revision,
		&decision.ID, &decisionKey, &decision.PeriodID,
		&decision.SourceOperationID, &decision.Current, &decision.Requested,
		&limitAmount, &unlimited, &decision.Allowed, &decision.SoftExceeded,
		&decision.DecidedAt)
	if storage.notFound(err) {
		return usageapp.ReservationReplay{}, false, nil
	}
	if err != nil {
		return usageapp.ReservationReplay{}, false,
			fmt.Errorf("find billing usage reservation: %w", err)
	}
	decision.EntitlementKey, err = entitlement.NewKey(decisionKey)
	if err == nil {
		decision.Limit, err = entitlementValue(limitAmount, unlimited)
	}
	if err == nil && hasReservation {
		reservation.EntitlementKey, err = entitlement.NewKey(reservationKey)
		reservation.State = usagedomain.ReservationState(state)
	} else if !hasReservation {
		reservation = usageapp.Reservation{}
	}
	if err != nil {
		return usageapp.ReservationReplay{}, false, usageapp.ErrSnapshot
	}
	return usageapp.ReservationReplay{Reservation: reservation, Decision: decision}, true, nil
}

func (storage *store) lockReservation(
	ctx context.Context,
	reservationID uuid.UUID,
) (usageapp.Reservation, bool, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return usageapp.Reservation{}, false, err
	}
	result := usageapp.Reservation{}
	var rawKey, state string
	err = storage.queryRow(ctx, lockReservationSQL, authorityID, storage.tenantID,
		reservationID,
	).Scan(&result.ID, &result.DecisionID, &rawKey, &result.PeriodID,
		&result.SourceOperationID, &result.Amount, &state, &result.ExpiresAt,
		&result.EffectStartedAt, &result.EffectReceiptID,
		&result.ResolutionReceipt, &result.CommittedAmount, &result.Revision)
	if storage.notFound(err) {
		return usageapp.Reservation{}, false, nil
	}
	if err != nil {
		return usageapp.Reservation{}, false,
			fmt.Errorf("lock billing usage reservation: %w", err)
	}
	result.EntitlementKey, err = entitlement.NewKey(rawKey)
	result.State = usagedomain.ReservationState(state)
	if err != nil {
		return usageapp.Reservation{}, false, usageapp.ErrSnapshot
	}
	return result, true, nil
}

func entitlementValue(limit *int64, unlimited bool) (entitlement.Value, error) {
	if unlimited {
		return entitlement.UnlimitedValue(), nil
	}
	if limit == nil {
		return entitlement.Value{}, entitlement.ErrValue
	}
	return entitlement.IntegerValue(*limit)
}
