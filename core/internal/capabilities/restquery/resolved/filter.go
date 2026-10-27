package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type Filter struct {
	kind      api.RequestFilterKind
	predicate *Predicate
	children  []Filter
	validated bool
}

type Predicate struct {
	field     api.FieldDefinition
	operator  api.Operator
	values    []Value
	path      RequestPath
	validated bool
}

func NewPredicate(
	field api.FieldDefinition, operator api.Operator, values []Value, path RequestPath,
) (Predicate, error) {
	predicate := Predicate{
		field: field, operator: operator, values: cloneValues(values),
		path: path, validated: true,
	}
	if err := predicate.Validate(); err != nil {
		return Predicate{}, err
	}
	return predicate, nil
}

func NewPredicateFilter(predicate Predicate) (Filter, error) {
	owned := predicate.clone()
	filter := Filter{
		kind: api.RequestFilterPredicate, predicate: &owned, validated: true,
	}
	if err := filter.Validate(); err != nil {
		return Filter{}, err
	}
	return filter, nil
}

func NewBooleanFilter(kind api.RequestFilterKind, children []Filter) (Filter, error) {
	filter := Filter{kind: kind, children: cloneFilters(children), validated: true}
	if err := filter.Validate(); err != nil {
		return Filter{}, err
	}
	return filter, nil
}

func (filter Filter) Validate() error {
	if !measureFilter(filter, 0).valid {
		return fmt.Errorf("resolved REST Query filter is invalid")
	}
	return nil
}

func (filter Filter) Kind() (api.RequestFilterKind, error) {
	if err := filter.Validate(); err != nil {
		return 0, err
	}
	return filter.kind, nil
}

func (filter Filter) Predicate() (Predicate, bool, error) {
	if err := filter.Validate(); err != nil {
		return Predicate{}, false, err
	}
	if filter.predicate == nil {
		return Predicate{}, false, nil
	}
	return filter.predicate.clone(), true, nil
}

func (filter Filter) Children() ([]Filter, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return cloneFilters(filter.children), nil
}

func (filter Filter) clone() Filter {
	result := filter
	result.children = cloneFilters(filter.children)
	if filter.predicate != nil {
		owned := filter.predicate.clone()
		result.predicate = &owned
	}
	return result
}

func cloneFilters(values []Filter) []Filter {
	if values == nil {
		return nil
	}
	result := make([]Filter, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}

func (Filter) String() string                        { return valueRedaction }
func (Filter) GoString() string                      { return valueRedaction }
func (filter Filter) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Filter) MarshalJSON() ([]byte, error)          { return rejectDirectJSONMarshal() }
