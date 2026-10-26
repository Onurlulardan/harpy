package discount

import (
	"bytes"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

type Application struct {
	identifier Identifier
	amount     money.Money
}

func (application Application) Identifier() Identifier { return application.identifier }
func (application Application) Amount() money.Money    { return application.amount }

type Result struct {
	net          money.Money
	discount     money.Money
	applications []Application
}

func (result Result) Net() money.Money      { return result.net }
func (result Result) Discount() money.Money { return result.discount }
func (result Result) Applications() []Application {
	return append([]Application(nil), result.applications...)
}

func Apply(net money.Money, cycleOrdinal int64, rules []Rule) (Result, error) {
	if net.MinorUnits() < 0 {
		return Result{}, ErrApply
	}
	if _, err := money.New(net.MinorUnits(), net.Currency(), net.CurrencyRegistryVersion()); err != nil {
		return Result{}, ErrApply
	}
	ordered := append([]Rule(nil), rules...)
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].priority != ordered[right].priority {
			return ordered[left].priority < ordered[right].priority
		}
		return bytes.Compare(ordered[left].id[:], ordered[right].id[:]) < 0
	})
	remaining := net
	applications := make([]Application, 0, len(ordered))
	seen := make(map[Identifier]struct{}, len(ordered))
	for _, rule := range ordered {
		if !validRule(rule) {
			return Result{}, ErrApply
		}
		if _, duplicate := seen[rule.id]; duplicate {
			return Result{}, ErrApply
		}
		seen[rule.id] = struct{}{}
		applies, err := rule.duration.Applies(cycleOrdinal)
		if err != nil {
			return Result{}, ErrApply
		}
		if !applies {
			continue
		}
		amount, err := ruleAmount(remaining, rule)
		if err != nil {
			return Result{}, ErrApply
		}
		remaining, err = money.Subtract(remaining, amount)
		if err != nil {
			return Result{}, ErrApply
		}
		applications = append(applications, Application{rule.id, amount})
	}
	discount, err := money.Subtract(net, remaining)
	if err != nil {
		return Result{}, ErrApply
	}
	return Result{remaining, discount, applications}, nil
}

func ruleAmount(remaining money.Money, rule Rule) (money.Money, error) {
	if rule.kind == BasisPoint {
		return money.ApplyBasisPoints(remaining, rule.rate)
	}
	if rule.fixed.Currency() != remaining.Currency() ||
		rule.fixed.CurrencyRegistryVersion() != remaining.CurrencyRegistryVersion() {
		return money.Money{}, ErrApply
	}
	if rule.fixed.MinorUnits() > remaining.MinorUnits() {
		return money.New(remaining.MinorUnits(), remaining.Currency(),
			remaining.CurrencyRegistryVersion())
	}
	return rule.fixed, nil
}

func validRule(rule Rule) bool {
	if zeroIdentifier(rule.id) || !rule.duration.valid() {
		return false
	}
	if rule.kind == BasisPoint {
		return rule.rate.Value() > 0 && rule.rate.Value() <= money.MaximumBasisPoints
	}
	if rule.kind != Fixed || rule.fixed.MinorUnits() <= 0 {
		return false
	}
	_, err := money.New(rule.fixed.MinorUnits(), rule.fixed.Currency(),
		rule.fixed.CurrencyRegistryVersion())
	return err == nil
}
