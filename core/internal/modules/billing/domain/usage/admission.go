package usage

import (
	"math"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

type Admission struct {
	allowed           bool
	softLimitExceeded bool
	projected         int64
}

func (admission Admission) Allowed() bool           { return admission.allowed }
func (admission Admission) SoftLimitExceeded() bool { return admission.softLimitExceeded }
func (admission Admission) Projected() int64        { return admission.projected }

func Admit(
	definition entitlement.Definition,
	limit entitlement.Value,
	committed int64,
	reserved int64,
	requested int64,
) (Admission, error) {
	if !validCounter(definition) || committed < 0 || reserved < 0 || requested <= 0 {
		return Admission{}, ErrAdmission
	}
	projected, ok := checkedSum(committed, reserved, requested)
	if !ok {
		return Admission{}, ErrAdmission
	}
	if limit.Unlimited() {
		if !definition.AllowsUnlimited() {
			return Admission{}, ErrAdmission
		}
		return Admission{allowed: true, projected: projected}, nil
	}
	maximum, ok := limit.Integer()
	if !ok {
		return Admission{}, ErrAdmission
	}
	exceeded := projected > maximum
	if !exceeded {
		return Admission{allowed: true, projected: projected}, nil
	}
	if definition.Enforcement() == entitlement.Hard {
		return Admission{projected: projected}, nil
	}
	return Admission{allowed: true, softLimitExceeded: true, projected: projected}, nil
}

func validCounter(definition entitlement.Definition) bool {
	return definition.Key().Valid() && definition.ValueKind() == entitlement.Integer &&
		definition.MeterKind() == entitlement.Counter &&
		(definition.Enforcement() == entitlement.Hard ||
			definition.Enforcement() == entitlement.Soft)
}

func checkedSum(values ...int64) (int64, bool) {
	var total int64
	for _, value := range values {
		if value > math.MaxInt64-total {
			return 0, false
		}
		total += value
	}
	return total, true
}
