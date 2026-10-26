package usage

import (
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	"github.com/google/uuid"
)

type currentRow struct {
	periodCount, overrideCount                          int64
	periodID                                            uuid.UUID
	key, valueKind, meterKind, resetPolicy, enforcement string
	allowsUnlimited                                     bool
	periodStart, periodEnd                              time.Time
	periodRevision                                      int64
	committed, reserved, gauge                          int64
	totalRevision                                       int64
	updatedAt                                           time.Time
	limitAmount                                         dbInt64
	unlimited, limitPresent                             bool
}

func scanCurrentRow(scanner interface{ Scan(...any) error }) (currentRow, error) {
	var value currentRow
	err := scanner.Scan(
		&value.periodCount, &value.overrideCount, &value.periodID, &value.key,
		&value.valueKind, &value.meterKind, &value.resetPolicy, &value.enforcement,
		&value.allowsUnlimited, &value.periodStart, &value.periodEnd,
		&value.periodRevision, &value.committed, &value.reserved, &value.gauge,
		&value.totalRevision, &value.updatedAt, &value.limitAmount,
		&value.unlimited, &value.limitPresent,
	)
	return value, err
}

func (value currentRow) snapshot() (usageapp.MeterSnapshot, error) {
	if value.periodCount != 1 || value.overrideCount > 1 || !value.limitPresent {
		return usageapp.MeterSnapshot{}, usageapp.ErrQueryInvariant
	}
	key, err := entitlement.NewKey(value.key)
	if err != nil {
		return usageapp.MeterSnapshot{}, usageapp.ErrQueryInvariant
	}
	definition, err := entitlement.NewDefinition(
		key, entitlement.ValueKind(value.valueKind), entitlement.MeterKind(value.meterKind),
		entitlement.ResetPolicy(value.resetPolicy), entitlement.Enforcement(value.enforcement),
		value.allowsUnlimited,
	)
	if err != nil {
		return usageapp.MeterSnapshot{}, usageapp.ErrQueryInvariant
	}
	limit := entitlement.UnlimitedValue()
	if !value.unlimited {
		if !value.limitAmount.Valid {
			return usageapp.MeterSnapshot{}, usageapp.ErrQueryInvariant
		}
		limit, err = entitlement.IntegerValue(value.limitAmount.Int64)
		if err != nil {
			return usageapp.MeterSnapshot{}, usageapp.ErrQueryInvariant
		}
	} else if value.limitAmount.Valid {
		return usageapp.MeterSnapshot{}, usageapp.ErrQueryInvariant
	}
	period, err := billingperiod.New(value.periodStart.UTC(), value.periodEnd.UTC())
	if err != nil {
		return usageapp.MeterSnapshot{}, usageapp.ErrQueryInvariant
	}
	return usageapp.NewMeterSnapshot(usageapp.MeterSnapshotInput{
		Definition: definition, Limit: limit, PeriodID: value.periodID, Period: period,
		Committed: value.committed, Reserved: value.reserved, Gauge: value.gauge,
		PeriodRevision: value.periodRevision, TotalRevision: value.totalRevision,
		UpdatedAt: value.updatedAt.UTC(),
	})
}
