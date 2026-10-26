package discount

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func (policy Policy) Code() Code            { return policy.code }
func (policy Policy) Kind() RuleKind        { return policy.kind }
func (policy Policy) Duration() Duration    { return policy.duration }
func (policy Policy) Priority() int32       { return policy.priority }
func (policy Policy) ValidStart() time.Time { return policy.validStart }
func (policy Policy) AllOfferings() bool    { return policy.allOfferings }
func (policy Policy) OfferingIDs() []Identifier {
	return append([]Identifier(nil), policy.offeringIDs...)
}

func (policy Policy) BasisPoints() (money.BasisPoints, bool) {
	return policy.basisPoints, policy.kind == BasisPoint
}

func (policy Policy) FixedAmount() (money.Money, bool) {
	return policy.fixedAmount, policy.kind == Fixed
}

func (policy Policy) GlobalLimit() *int64 { return copyLimit(policy.globalLimit) }
func (policy Policy) TenantLimit() *int64 { return copyLimit(policy.tenantLimit) }

func (policy Policy) ValidEnd() *time.Time {
	if policy.validEnd == nil {
		return nil
	}
	end := *policy.validEnd
	return &end
}

func (policy Policy) Rule(identifier Identifier) (Rule, error) {
	if policy.kind == BasisPoint {
		return NewBasisPointRule(
			identifier, policy.priority, policy.basisPoints, policy.duration,
		)
	}
	if policy.kind == Fixed {
		return NewFixedRule(
			identifier, policy.priority, policy.fixedAmount, policy.duration,
		)
	}
	return Rule{}, ErrPolicy
}

func (policy Policy) Valid() bool {
	if !policy.code.Valid() || !policy.duration.Valid() || policy.priority < 0 ||
		!validLimit(policy.globalLimit) || !validLimit(policy.tenantLimit) ||
		!validApplicability(policy.allOfferings, policy.offeringIDs) {
		return false
	}
	start, err := canonicalInstant(policy.validStart)
	if err != nil {
		return false
	}
	if _, err = canonicalPolicyEnd(policy.validEnd, start); err != nil {
		return false
	}
	_, err = policy.Rule(Identifier{15: 1})
	return err == nil
}
