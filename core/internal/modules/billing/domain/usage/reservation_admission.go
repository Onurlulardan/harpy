package usage

import "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"

// AdmitReservation applies the entitlement's meter semantics while preserving
// pending reservations as part of the hard-limit calculation.
func AdmitReservation(definition entitlement.Definition, limit entitlement.Value,
	committed, reserved, gauge, requested int64,
) (Admission, error) {
	switch definition.MeterKind() {
	case entitlement.Counter:
		if gauge != 0 {
			return Admission{}, ErrAdmission
		}
		return Admit(definition, limit, committed, reserved, requested)
	case entitlement.Gauge:
		if committed != 0 || reserved < 0 || gauge < 0 || requested <= 0 {
			return Admission{}, ErrAdmission
		}
		current, ok := checkedSum(gauge, reserved)
		if !ok {
			return Admission{}, ErrAdmission
		}
		adjustment, err := AdjustGauge(definition, limit, current, requested)
		if err != nil {
			return Admission{}, ErrAdmission
		}
		return Admission{
			allowed:           adjustment.Allowed(),
			softLimitExceeded: adjustment.SoftLimitExceeded(),
			projected:         adjustment.After(),
		}, nil
	default:
		return Admission{}, ErrAdmission
	}
}
