package api

import "fmt"

const maximumAggregateSorts = 3

type RequestAggregateSort struct {
	reference AggregateReference
	name      string
	direction Direction
}

func NewRequestAggregateSort(
	reference AggregateReference, name string, direction Direction,
) (RequestAggregateSort, error) {
	sort := RequestAggregateSort{reference: reference, name: name, direction: direction}
	if !sort.valid() {
		return RequestAggregateSort{}, fmt.Errorf("REST Query aggregate sort is invalid")
	}
	return sort, nil
}

func (sort RequestAggregateSort) Reference() AggregateReference { return sort.reference }
func (sort RequestAggregateSort) Name() string                  { return sort.name }
func (sort RequestAggregateSort) Direction() Direction          { return sort.direction }
func (sort RequestAggregateSort) String() string                { return queryValueRedaction }
func (sort RequestAggregateSort) GoString() string              { return queryValueRedaction }
func (sort RequestAggregateSort) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestAggregateSort) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}

func (sort RequestAggregateSort) valid() bool {
	return sort.direction.valid() &&
		validAggregateReferenceName(sort.reference, sort.name)
}

func cloneRequestAggregateSorts(
	values []RequestAggregateSort,
	groups map[string]struct{},
	measures map[string]struct{},
) ([]RequestAggregateSort, error) {
	if len(values) > maximumAggregateSorts {
		return nil, fmt.Errorf("REST Query aggregate sort count exceeds its limit")
	}
	result := append([]RequestAggregateSort{}, values...)
	type sortKey struct {
		reference AggregateReference
		name      string
	}
	seen := make(map[sortKey]struct{}, len(result))
	for index, value := range result {
		if !value.valid() || !aggregateReferenceDeclared(
			value.reference, value.name, groups, measures,
		) {
			return nil, fmt.Errorf("REST Query aggregate sort %d is invalid", index)
		}
		key := sortKey{reference: value.reference, name: value.name}
		if _, duplicate := seen[key]; duplicate {
			return nil, fmt.Errorf("REST Query aggregate sort reference is duplicated")
		}
		seen[key] = struct{}{}
	}
	return result, nil
}

func aggregateReferenceDeclared(
	reference AggregateReference,
	name string,
	groups map[string]struct{},
	measures map[string]struct{},
) bool {
	if reference == AggregateReferenceGroup {
		_, found := groups[name]
		return found
	}
	if reference == AggregateReferenceMeasure {
		_, found := measures[name]
		return found
	}
	return false
}
