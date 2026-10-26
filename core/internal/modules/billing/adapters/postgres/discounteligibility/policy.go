package discounteligibility

import (
	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func (record eligibilityRecord) policy(
	offeringIDs []discountdomain.Identifier,
) (discountdomain.Policy, error) {
	code, err := discountdomain.NormalizeCode(record.code)
	if err != nil || record.expectedOfferingCount != int16(len(offeringIDs)) ||
		record.priority < 0 {
		return discountdomain.Policy{}, discountapp.ErrQueryInvariant
	}
	duration, err := record.discountDuration()
	if err != nil {
		return discountdomain.Policy{}, discountapp.ErrQueryInvariant
	}
	switch record.kind {
	case "basis_point":
		if record.basisPoints == nil || record.fixedMinorUnits != nil ||
			record.fixedCurrency != nil || record.registryVersion != nil {
			return discountdomain.Policy{}, discountapp.ErrQueryInvariant
		}
		rate, rateErr := money.NewBasisPoints(int64(*record.basisPoints))
		if rateErr != nil {
			return discountdomain.Policy{}, discountapp.ErrQueryInvariant
		}
		return discountdomain.NewBasisPointPolicy(
			code, rate, duration, record.priority, record.globalLimit,
			record.tenantLimit, record.validStart, record.validEnd,
			record.allOfferings, offeringIDs,
		)
	case "fixed":
		if record.fixedMinorUnits == nil || record.fixedCurrency == nil ||
			record.registryVersion == nil || record.basisPoints != nil {
			return discountdomain.Policy{}, discountapp.ErrQueryInvariant
		}
		amount, amountErr := money.New(
			*record.fixedMinorUnits, *record.fixedCurrency, *record.registryVersion,
		)
		if amountErr != nil {
			return discountdomain.Policy{}, discountapp.ErrQueryInvariant
		}
		return discountdomain.NewFixedPolicy(
			code, amount, duration, record.priority, record.globalLimit,
			record.tenantLimit, record.validStart, record.validEnd,
			record.allOfferings, offeringIDs,
		)
	default:
		return discountdomain.Policy{}, discountapp.ErrQueryInvariant
	}
}

func (record eligibilityRecord) discountDuration() (discountdomain.Duration, error) {
	switch record.duration {
	case "once":
		if record.repeatingCycles == nil {
			return discountdomain.OnceDuration(), nil
		}
	case "forever":
		if record.repeatingCycles == nil {
			return discountdomain.ForeverDuration(), nil
		}
	case "repeating":
		if record.repeatingCycles != nil {
			return discountdomain.RepeatingDuration(*record.repeatingCycles)
		}
	}
	return discountdomain.Duration{}, discountapp.ErrQueryInvariant
}
