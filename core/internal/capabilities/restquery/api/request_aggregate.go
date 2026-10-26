package api

import "fmt"

const (
	maximumAggregateGroups = 3
	defaultAggregateLimit  = 100
	maximumAggregateLimit  = 500
)

type RequestAggregateInput struct {
	GroupBy  []string
	Measures []RequestMeasure
	Having   *RequestHaving
	Sort     []RequestAggregateSort
	Limit    *int
}

type RequestAggregate struct {
	groupBy  []string
	measures []RequestMeasure
	having   *RequestHaving
	sorts    []RequestAggregateSort
	limit    int
}

func NewRequestAggregate(input RequestAggregateInput) (RequestAggregate, error) {
	groupBy, err := cloneAggregateGroups(input.GroupBy)
	if err != nil {
		return RequestAggregate{}, err
	}
	measures, err := cloneRequestMeasures(input.Measures)
	if err != nil {
		return RequestAggregate{}, err
	}
	groups, measureNames := aggregateDeclarations(groupBy, measures)
	having, err := cloneOptionalRequestHaving(input.Having, groups, measureNames)
	if err != nil {
		return RequestAggregate{}, err
	}
	sorts, err := cloneRequestAggregateSorts(input.Sort, groups, measureNames)
	if err != nil {
		return RequestAggregate{}, err
	}
	limit := defaultAggregateLimit
	if input.Limit != nil {
		limit = *input.Limit
	}
	result := RequestAggregate{
		groupBy: groupBy, measures: measures, having: having,
		sorts: sorts, limit: limit,
	}
	if err := result.Validate(); err != nil {
		return RequestAggregate{}, err
	}
	return result, nil
}

func (aggregate RequestAggregate) GroupBy() []string {
	return append([]string{}, aggregate.groupBy...)
}
func (aggregate RequestAggregate) Measures() []RequestMeasure {
	return append([]RequestMeasure{}, aggregate.measures...)
}
func (aggregate RequestAggregate) Having() (RequestHaving, bool) {
	if aggregate.having == nil {
		return RequestHaving{}, false
	}
	return aggregate.having.clone(), true
}
func (aggregate RequestAggregate) Sort() []RequestAggregateSort {
	return append([]RequestAggregateSort{}, aggregate.sorts...)
}
func (aggregate RequestAggregate) Limit() int       { return aggregate.limit }
func (aggregate RequestAggregate) String() string   { return queryValueRedaction }
func (aggregate RequestAggregate) GoString() string { return queryValueRedaction }
func (aggregate RequestAggregate) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestAggregate) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
func (RequestAggregateInput) String() string          { return queryValueRedaction }
func (RequestAggregateInput) GoString() string        { return queryValueRedaction }
func (RequestAggregateInput) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestAggregateInput) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}

func (aggregate RequestAggregate) clone() RequestAggregate {
	result := aggregate
	result.groupBy = append([]string{}, aggregate.groupBy...)
	result.measures = append([]RequestMeasure{}, aggregate.measures...)
	result.sorts = append([]RequestAggregateSort{}, aggregate.sorts...)
	if aggregate.having != nil {
		having := aggregate.having.clone()
		result.having = &having
	}
	return result
}
