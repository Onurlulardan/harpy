package discount

import (
	"sort"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

type Policy struct {
	code         Code
	kind         RuleKind
	basisPoints  money.BasisPoints
	fixedAmount  money.Money
	duration     Duration
	priority     int32
	globalLimit  *int64
	tenantLimit  *int64
	validStart   time.Time
	validEnd     *time.Time
	allOfferings bool
	offeringIDs  []Identifier
}

func NewBasisPointPolicy(
	code Code,
	basisPoints money.BasisPoints,
	duration Duration,
	priority int32,
	globalLimit *int64,
	tenantLimit *int64,
	validStart time.Time,
	validEnd *time.Time,
	allOfferings bool,
	offeringIDs []Identifier,
) (Policy, error) {
	if basisPoints.Value() <= 0 || basisPoints.Value() > money.MaximumBasisPoints {
		return Policy{}, ErrPolicy
	}
	return newPolicy(code, BasisPoint, basisPoints, money.Money{}, duration, priority,
		globalLimit, tenantLimit, validStart, validEnd, allOfferings, offeringIDs)
}

func NewFixedPolicy(
	code Code,
	fixedAmount money.Money,
	duration Duration,
	priority int32,
	globalLimit *int64,
	tenantLimit *int64,
	validStart time.Time,
	validEnd *time.Time,
	allOfferings bool,
	offeringIDs []Identifier,
) (Policy, error) {
	if fixedAmount.MinorUnits() <= 0 {
		return Policy{}, ErrPolicy
	}
	verified, err := money.New(fixedAmount.MinorUnits(), fixedAmount.Currency(),
		fixedAmount.CurrencyRegistryVersion())
	if err != nil || verified.Exponent() != fixedAmount.Exponent() {
		return Policy{}, ErrPolicy
	}
	return newPolicy(code, Fixed, 0, fixedAmount, duration, priority,
		globalLimit, tenantLimit, validStart, validEnd, allOfferings, offeringIDs)
}

func newPolicy(
	code Code, kind RuleKind, basisPoints money.BasisPoints, fixedAmount money.Money,
	duration Duration, priority int32, globalLimit *int64, tenantLimit *int64,
	validStart time.Time, validEnd *time.Time, allOfferings bool,
	offeringIDs []Identifier,
) (Policy, error) {
	start, err := canonicalInstant(validStart)
	if err != nil || !code.Valid() || !duration.Valid() || priority < 0 ||
		!validLimit(globalLimit) || !validLimit(tenantLimit) ||
		!validApplicability(allOfferings, offeringIDs) {
		return Policy{}, ErrPolicy
	}
	end, err := canonicalPolicyEnd(validEnd, start)
	if err != nil {
		return Policy{}, ErrPolicy
	}
	ids := append([]Identifier(nil), offeringIDs...)
	sort.Slice(ids, func(left, right int) bool {
		return compareIdentifier(ids[left], ids[right]) < 0
	})
	return Policy{code: code, kind: kind, basisPoints: basisPoints,
		fixedAmount: fixedAmount, duration: duration, priority: priority,
		globalLimit: copyLimit(globalLimit), tenantLimit: copyLimit(tenantLimit),
		validStart: start, validEnd: end, allOfferings: allOfferings,
		offeringIDs: ids}, nil
}

func validLimit(value *int64) bool { return value == nil || *value > 0 }

func copyLimit(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
