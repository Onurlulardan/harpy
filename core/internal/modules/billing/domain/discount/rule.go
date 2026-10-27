package discount

import "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"

type RuleKind string
type Identifier [16]byte

const (
	BasisPoint RuleKind = "basis_point"
	Fixed      RuleKind = "fixed"
)

type Rule struct {
	id       Identifier
	priority int32
	kind     RuleKind
	rate     money.BasisPoints
	fixed    money.Money
	duration Duration
}

func NewBasisPointRule(
	id Identifier,
	priority int32,
	rate money.BasisPoints,
	duration Duration,
) (Rule, error) {
	if zeroIdentifier(id) || rate.Value() <= 0 || rate.Value() > money.MaximumBasisPoints ||
		!duration.valid() {
		return Rule{}, ErrRule
	}
	return Rule{id: id, priority: priority, kind: BasisPoint, rate: rate,
		duration: duration}, nil
}

func NewFixedRule(
	id Identifier,
	priority int32,
	fixed money.Money,
	duration Duration,
) (Rule, error) {
	if zeroIdentifier(id) || fixed.MinorUnits() <= 0 || !duration.valid() {
		return Rule{}, ErrRule
	}
	if _, err := money.New(
		fixed.MinorUnits(), fixed.Currency(), fixed.CurrencyRegistryVersion()); err != nil {
		return Rule{}, ErrRule
	}
	return Rule{id: id, priority: priority, kind: Fixed, fixed: fixed,
		duration: duration}, nil
}

func (rule Rule) Identifier() Identifier { return rule.id }
func (rule Rule) Priority() int32        { return rule.priority }
func (rule Rule) Kind() RuleKind         { return rule.kind }
func (rule Rule) Duration() Duration     { return rule.duration }

func (rule Rule) BasisPoints() (money.BasisPoints, bool) {
	return rule.rate, rule.kind == BasisPoint
}

func (rule Rule) FixedAmount() (money.Money, bool) {
	return rule.fixed, rule.kind == Fixed
}

func zeroIdentifier(identifier Identifier) bool {
	return identifier == Identifier{}
}
