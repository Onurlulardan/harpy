package usage

import (
	"math"
	"slices"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	"github.com/google/uuid"
)

type MeterSnapshotInput struct {
	Definition                    entitlement.Definition
	Limit                         entitlement.Value
	PeriodID                      uuid.UUID
	Period                        billingperiod.Period
	Committed, Reserved, Gauge    int64
	PeriodRevision, TotalRevision int64
	UpdatedAt                     time.Time
}

type MeterSnapshot struct {
	definition                    entitlement.Definition
	limit                         entitlement.Value
	periodID                      uuid.UUID
	period                        billingperiod.Period
	committed, reserved, gauge    int64
	periodRevision, totalRevision int64
	updatedAt                     time.Time
}

type UsageSnapshot struct {
	tenantID uuid.UUID
	meters   []MeterSnapshot
}

func NewMeterSnapshot(input MeterSnapshotInput) (MeterSnapshot, error) {
	value := MeterSnapshot{
		definition: input.Definition, limit: input.Limit,
		periodID: input.PeriodID, period: input.Period,
		committed: input.Committed, reserved: input.Reserved, gauge: input.Gauge,
		periodRevision: input.PeriodRevision, totalRevision: input.TotalRevision,
		updatedAt: input.UpdatedAt,
	}
	if !value.valid() {
		return MeterSnapshot{}, ErrQueryInvariant
	}
	return value, nil
}

func NewUsageSnapshot(tenantID uuid.UUID, meters []MeterSnapshot) (UsageSnapshot, error) {
	if tenantID == uuid.Nil {
		return UsageSnapshot{}, ErrQueryInvariant
	}
	copyMeters := append([]MeterSnapshot(nil), meters...)
	slices.SortFunc(copyMeters, func(left, right MeterSnapshot) int {
		if left.Key() < right.Key() {
			return -1
		}
		if left.Key() > right.Key() {
			return 1
		}
		if left.periodID.String() < right.periodID.String() {
			return -1
		}
		if left.periodID.String() > right.periodID.String() {
			return 1
		}
		return 0
	})
	for index, meter := range copyMeters {
		if !meter.valid() || index > 0 && copyMeters[index-1].Key() == meter.Key() {
			return UsageSnapshot{}, ErrQueryInvariant
		}
	}
	return UsageSnapshot{tenantID: tenantID, meters: copyMeters}, nil
}

func (value UsageSnapshot) TenantID() uuid.UUID { return value.tenantID }
func (value UsageSnapshot) Meters() []MeterSnapshot {
	return append([]MeterSnapshot(nil), value.meters...)
}
func (value MeterSnapshot) Definition() entitlement.Definition { return value.definition }
func (value MeterSnapshot) Key() entitlement.Key               { return value.definition.Key() }
func (value MeterSnapshot) Limit() entitlement.Value           { return value.limit }
func (value MeterSnapshot) PeriodID() uuid.UUID                { return value.periodID }
func (value MeterSnapshot) Period() billingperiod.Period       { return value.period }
func (value MeterSnapshot) Committed() int64                   { return value.committed }
func (value MeterSnapshot) Reserved() int64                    { return value.reserved }
func (value MeterSnapshot) Gauge() int64                       { return value.gauge }
func (value MeterSnapshot) PeriodRevision() int64              { return value.periodRevision }
func (value MeterSnapshot) TotalRevision() int64               { return value.totalRevision }
func (value MeterSnapshot) UpdatedAt() time.Time               { return value.updatedAt }

func (value MeterSnapshot) Current() int64 {
	if value.definition.MeterKind() == entitlement.Counter {
		return value.committed + value.reserved
	}
	return value.gauge + value.reserved
}

func (value MeterSnapshot) valid() bool {
	if !validV7(value.periodID) || value.definition.ValueKind() != entitlement.Integer ||
		!value.definition.Accepts(value.limit) || value.periodRevision < 1 ||
		value.totalRevision < 1 || !validInstant(value.updatedAt) ||
		value.period.Start().Location() != time.UTC || value.period.End().Location() != time.UTC ||
		value.committed < 0 || value.reserved < 0 || value.gauge < 0 {
		return false
	}
	switch value.definition.MeterKind() {
	case entitlement.Counter:
		return value.gauge == 0 && value.committed <= math.MaxInt64-value.reserved
	case entitlement.Gauge:
		return value.committed == 0 && value.gauge <= math.MaxInt64-value.reserved
	default:
		return false
	}
}
