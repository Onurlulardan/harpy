package usage

import (
	"math"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

type GaugeAdjustment struct {
	allowed           bool
	softLimitExceeded bool
	before            int64
	after             int64
}

func (adjustment GaugeAdjustment) Allowed() bool           { return adjustment.allowed }
func (adjustment GaugeAdjustment) SoftLimitExceeded() bool { return adjustment.softLimitExceeded }
func (adjustment GaugeAdjustment) Before() int64           { return adjustment.before }
func (adjustment GaugeAdjustment) After() int64            { return adjustment.after }

func AdjustGauge(
	definition entitlement.Definition,
	limit entitlement.Value,
	current int64,
	delta int64,
) (GaugeAdjustment, error) {
	if !validGauge(definition) || current < 0 || delta == 0 {
		return GaugeAdjustment{}, ErrGauge
	}
	after, ok := checkedDelta(current, delta)
	if !ok || after < 0 {
		return GaugeAdjustment{}, ErrGauge
	}
	if delta < 0 {
		return GaugeAdjustment{allowed: true, before: current, after: after}, nil
	}
	if limit.Unlimited() {
		if !definition.AllowsUnlimited() {
			return GaugeAdjustment{}, ErrGauge
		}
		return GaugeAdjustment{allowed: true, before: current, after: after}, nil
	}
	maximum, ok := limit.Integer()
	if !ok {
		return GaugeAdjustment{}, ErrGauge
	}
	if after <= maximum {
		return GaugeAdjustment{allowed: true, before: current, after: after}, nil
	}
	if definition.Enforcement() == entitlement.Hard {
		return GaugeAdjustment{before: current, after: after}, nil
	}
	return GaugeAdjustment{
		allowed: true, softLimitExceeded: true, before: current, after: after,
	}, nil
}

func validGauge(definition entitlement.Definition) bool {
	return definition.Key().Valid() && definition.ValueKind() == entitlement.Integer &&
		definition.MeterKind() == entitlement.Gauge &&
		definition.ResetPolicy() == entitlement.Never &&
		(definition.Enforcement() == entitlement.Hard ||
			definition.Enforcement() == entitlement.Soft)
}

func checkedDelta(current, delta int64) (int64, bool) {
	if delta > 0 && current > math.MaxInt64-delta ||
		delta < 0 && current < math.MinInt64-delta {
		return 0, false
	}
	return current + delta, true
}
