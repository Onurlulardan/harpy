package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type Having struct {
	kind      api.RequestHavingKind
	predicate *HavingPredicate
	children  []Having
	validated bool
}

type HavingPredicate struct {
	reference AggregateReference
	operator  api.Operator
	values    []Value
	validated bool
}

func NewHavingPredicate(
	reference AggregateReference, operator api.Operator, values []Value,
) (HavingPredicate, error) {
	predicate := HavingPredicate{
		reference: reference, operator: operator,
		values: cloneValues(values), validated: true,
	}
	if err := predicate.Validate(); err != nil {
		return HavingPredicate{}, err
	}
	return predicate, nil
}

func NewPredicateHaving(predicate HavingPredicate) (Having, error) {
	owned := predicate.clone()
	having := Having{
		kind:      api.RequestHavingKindPredicate,
		predicate: &owned, validated: true,
	}
	if err := having.Validate(); err != nil {
		return Having{}, err
	}
	return having, nil
}

func NewBooleanHaving(
	kind api.RequestHavingKind, children []Having,
) (Having, error) {
	having := Having{kind: kind, children: cloneHavings(children), validated: true}
	if err := having.Validate(); err != nil {
		return Having{}, err
	}
	return having, nil
}

func (predicate HavingPredicate) Validate() error {
	if !predicate.validated || predicate.reference.Validate() != nil || predicate.values == nil {
		return fmt.Errorf("resolved REST Query HAVING predicate is invalid")
	}
	arity, valid := predicate.operator.Arity()
	if !valid || !resolvedArityValid(arity, len(predicate.values)) {
		return fmt.Errorf("resolved REST Query HAVING predicate is invalid")
	}
	for _, value := range predicate.values {
		if value.Validate() != nil {
			return fmt.Errorf("resolved REST Query HAVING predicate is invalid")
		}
	}
	return nil
}

func (having Having) Validate() error {
	if !measureHaving(having, 0).valid {
		return fmt.Errorf("resolved REST Query HAVING is invalid")
	}
	return nil
}

func (having Having) clone() Having {
	result := having
	result.children = cloneHavings(having.children)
	if having.predicate != nil {
		owned := having.predicate.clone()
		result.predicate = &owned
	}
	return result
}

func cloneHavings(values []Having) []Having {
	if values == nil {
		return nil
	}
	result := make([]Having, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}

func (predicate HavingPredicate) clone() HavingPredicate {
	result := predicate
	result.values = cloneValues(predicate.values)
	return result
}

func (Having) String() string                                    { return valueRedaction }
func (Having) GoString() string                                  { return valueRedaction }
func (having Having) Format(state fmt.State, _ rune)             { writeRedaction(state) }
func (Having) MarshalJSON() ([]byte, error)                      { return rejectDirectJSONMarshal() }
func (HavingPredicate) String() string                           { return valueRedaction }
func (HavingPredicate) GoString() string                         { return valueRedaction }
func (predicate HavingPredicate) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (HavingPredicate) MarshalJSON() ([]byte, error)             { return rejectDirectJSONMarshal() }
