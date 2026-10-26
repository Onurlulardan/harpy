package usage

import (
	"context"
	"fmt"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	"github.com/google/uuid"
)

type usageSnapshot struct {
	definition entitlement.Definition
	limit      entitlement.Value
	periodID   uuid.UUID
	period     billingperiod.Period
	committed  int64
	reserved   int64
	gauge      int64
	revision   int64
}

func (storage *store) lockSnapshot(
	ctx context.Context,
	key entitlement.Key,
	periodID uuid.UUID,
	meterKind string,
) (usageSnapshot, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return usageSnapshot{}, err
	}
	var rawKey, valueKind, meter, reset, enforcement string
	var allowsUnlimited, unlimited bool
	var limitAmount *int64
	var start, end time.Time
	result := usageSnapshot{}
	err = storage.queryRow(ctx, lockSnapshotSQL, authorityID, storage.tenantID,
		string(key), periodID, meterKind,
	).Scan(&rawKey, &valueKind, &meter, &reset, &enforcement,
		&allowsUnlimited, &limitAmount, &unlimited, &result.periodID,
		&start, &end, &result.committed, &result.reserved, &result.gauge,
		&result.revision)
	if err != nil {
		return usageSnapshot{}, fmt.Errorf("lock billing usage snapshot: %w", err)
	}
	normalizedKey, err := entitlement.NewKey(rawKey)
	if err != nil {
		return usageSnapshot{}, usageapp.ErrSnapshot
	}
	result.definition, err = entitlement.NewDefinition(
		normalizedKey, entitlement.ValueKind(valueKind), entitlement.MeterKind(meter),
		entitlement.ResetPolicy(reset), entitlement.Enforcement(enforcement),
		allowsUnlimited,
	)
	if err != nil {
		return usageSnapshot{}, usageapp.ErrSnapshot
	}
	if unlimited {
		result.limit = entitlement.UnlimitedValue()
	} else if limitAmount != nil {
		result.limit, err = entitlement.IntegerValue(*limitAmount)
	} else {
		err = entitlement.ErrValue
	}
	if err != nil || !result.definition.Accepts(result.limit) {
		return usageSnapshot{}, usageapp.ErrSnapshot
	}
	result.period, err = billingperiod.New(start, end)
	if err != nil {
		return usageSnapshot{}, usageapp.ErrSnapshot
	}
	return result, nil
}

func (storage *store) lockReservationSnapshot(
	ctx context.Context, key entitlement.Key, periodID uuid.UUID,
) (usageapp.ReservationSnapshot, error) {
	snapshot, err := storage.lockSnapshot(ctx, key, periodID, "reservation")
	if err != nil {
		return usageapp.ReservationSnapshot{}, err
	}
	return usageapp.ReservationSnapshot{Definition: snapshot.definition, Limit: snapshot.limit,
		PeriodID: snapshot.periodID, Period: snapshot.period,
		Committed: snapshot.committed, Reserved: snapshot.reserved, Gauge: snapshot.gauge,
		Revision: snapshot.revision}, nil
}

func (storage *store) lockGauge(
	ctx context.Context, key entitlement.Key, periodID uuid.UUID,
) (usageapp.GaugeSnapshot, error) {
	snapshot, err := storage.lockSnapshot(ctx, key, periodID, string(entitlement.Gauge))
	if err != nil {
		return usageapp.GaugeSnapshot{}, err
	}
	return usageapp.GaugeSnapshot{Definition: snapshot.definition, Limit: snapshot.limit,
		PeriodID: snapshot.periodID, Period: snapshot.period,
		Current: snapshot.gauge, Revision: snapshot.revision}, nil
}
