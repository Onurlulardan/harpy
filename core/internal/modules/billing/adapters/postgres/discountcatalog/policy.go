package discountcatalog

import discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"

type policyProjection struct {
	kind, currency, duration string
	fixedMinor, registry     any
	basisPoints, cycles      any
}

func projectPolicy(policy discountdomain.Policy) policyProjection {
	projection := policyProjection{
		kind: string(policy.Kind()), duration: string(policy.Duration().Kind()),
	}
	if fixed, present := policy.FixedAmount(); present {
		projection.fixedMinor = fixed.MinorUnits()
		projection.currency = fixed.Currency()
		projection.registry = fixed.CurrencyRegistryVersion()
	}
	if basisPoints, present := policy.BasisPoints(); present {
		projection.basisPoints = basisPoints.Value()
	}
	if policy.Duration().Kind() == discountdomain.Repeating {
		projection.cycles = policy.Duration().Cycles()
	}
	return projection
}
