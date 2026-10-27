package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type MandatoryFilter struct {
	kind      api.RequestFilterKind
	predicate *MandatoryPredicate
	children  []MandatoryFilter
	validated bool
}

type MandatoryPredicate struct {
	field     api.FieldDefinition
	operator  api.Operator
	values    []Value
	validated bool
}

func NewMandatoryPredicate(
	field api.FieldDefinition,
	operator api.Operator,
	values []Value,
) (MandatoryPredicate, error) {
	predicate := MandatoryPredicate{
		field: field, operator: operator,
		values: cloneValues(values), validated: true,
	}
	if err := predicate.Validate(); err != nil {
		return MandatoryPredicate{}, err
	}
	return predicate, nil
}

func NewMandatoryPredicateFilter(
	predicate MandatoryPredicate,
) (MandatoryFilter, error) {
	owned := predicate.clone()
	filter := MandatoryFilter{
		kind: api.RequestFilterPredicate, predicate: &owned, validated: true,
	}
	if err := filter.Validate(); err != nil {
		return MandatoryFilter{}, err
	}
	return filter, nil
}

func NewMandatoryAnd(children []MandatoryFilter) (MandatoryFilter, error) {
	return newMandatoryBoolean(api.RequestFilterAnd, children)
}

func NewMandatoryOr(children []MandatoryFilter) (MandatoryFilter, error) {
	return newMandatoryBoolean(api.RequestFilterOr, children)
}

func NewMandatoryNot(child MandatoryFilter) (MandatoryFilter, error) {
	return newMandatoryBoolean(api.RequestFilterNot, []MandatoryFilter{child})
}

func newMandatoryBoolean(
	kind api.RequestFilterKind,
	children []MandatoryFilter,
) (MandatoryFilter, error) {
	filter := MandatoryFilter{
		kind: kind, children: cloneMandatoryFilters(children), validated: true,
	}
	if err := filter.Validate(); err != nil {
		return MandatoryFilter{}, err
	}
	return filter, nil
}

func (MandatoryFilter) String() string   { return valueRedaction }
func (MandatoryFilter) GoString() string { return valueRedaction }
func (filter MandatoryFilter) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (MandatoryFilter) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (MandatoryPredicate) String() string   { return valueRedaction }
func (MandatoryPredicate) GoString() string { return valueRedaction }
func (predicate MandatoryPredicate) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (MandatoryPredicate) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
