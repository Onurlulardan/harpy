package api

import "fmt"

type AggregateQueryRequestInput struct {
	Where     *RequestFilter
	Search    *RequestSearch
	Aggregate RequestAggregate
}

func NewAggregateQueryRequest(input AggregateQueryRequestInput) (QueryRequest, error) {
	where, err := cloneOptionalRequestFilter(input.Where)
	if err != nil {
		return QueryRequest{}, err
	}
	search, err := cloneOptionalRequestSearch(input.Search)
	if err != nil {
		return QueryRequest{}, err
	}
	if err := input.Aggregate.Validate(); err != nil {
		return QueryRequest{}, err
	}
	aggregate := input.Aggregate.clone()
	request := QueryRequest{
		mode: QueryModeAggregate, where: where, search: search,
		aggregate: &aggregate,
	}
	if err := request.Validate(); err != nil {
		return QueryRequest{}, err
	}
	return request, nil
}

func (input AggregateQueryRequestInput) String() string   { return queryValueRedaction }
func (input AggregateQueryRequestInput) GoString() string { return queryValueRedaction }
func (input AggregateQueryRequestInput) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (AggregateQueryRequestInput) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
