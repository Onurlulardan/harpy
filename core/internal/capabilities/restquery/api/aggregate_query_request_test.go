package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewAggregateQueryRequestSupportsPreAggregateWhereAndSearch(t *testing.T) {
	where := mustPredicate(t)
	search := mustRequestSearch(t, "term")
	aggregate, err := NewRequestAggregate(RequestAggregateInput{
		Measures: []RequestMeasure{
			mustRequestMeasure(t, "row_count", AggregateFunctionCount, nil),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	input := AggregateQueryRequestInput{Where: &where, Search: &search, Aggregate: aggregate}
	request, err := NewAggregateQueryRequest(input)
	if err != nil {
		t.Fatal(err)
	}
	where = RequestFilter{}
	search = RequestSearch{}
	input.Aggregate.measures[0] = RequestMeasure{}
	stored, found := request.Aggregate()
	if request.Mode() != QueryModeAggregate || !found || stored.Measures()[0].Name() != "row_count" {
		t.Fatalf("unexpected aggregate query: %#v", request)
	}
	if _, found := request.Where(); !found {
		t.Fatal("aggregate query lost root where")
	}
	if value, found := request.Search(); !found || value.Term() != "term" {
		t.Fatal("aggregate query lost root search")
	}
	stored.measures[0] = RequestMeasure{}
	if request.aggregate.measures[0].Name() != "row_count" {
		t.Fatal("aggregate accessor exposed mutable state")
	}
}

func TestAggregateQueryRequestIsRedactedAndCannotMarshal(t *testing.T) {
	aggregate, _ := NewRequestAggregate(RequestAggregateInput{Measures: []RequestMeasure{
		mustRequestMeasure(t, "row_count", AggregateFunctionCount, nil),
	}})
	input := AggregateQueryRequestInput{Aggregate: aggregate}
	request, err := NewAggregateQueryRequest(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{
		fmt.Sprint(input), fmt.Sprintf("%#v", input),
		fmt.Sprint(request), fmt.Sprintf("%#v", request),
	} {
		if value != queryValueRedaction {
			t.Fatalf("aggregate query formatting leaked: %q", value)
		}
	}
	for _, value := range []any{input, request} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}

func TestAggregateQueryRequestValidateRejectsRecordStateAndMalformedAggregate(t *testing.T) {
	aggregate, _ := NewRequestAggregate(RequestAggregateInput{Measures: []RequestMeasure{
		mustRequestMeasure(t, "row_count", AggregateFunctionCount, nil),
	}})
	request, err := NewAggregateQueryRequest(AggregateQueryRequestInput{Aggregate: aggregate})
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*QueryRequest){
		"select":    func(value *QueryRequest) { value.selectFields = []string{} },
		"sort":      func(value *QueryRequest) { value.sorts = []RequestSort{} },
		"page":      func(value *QueryRequest) { value.page = mustRequestPage(t, 1, nil) },
		"include":   func(value *QueryRequest) { value.includes = []RequestInclude{} },
		"count":     func(value *QueryRequest) { value.count = CountModeNone },
		"aggregate": func(value *QueryRequest) { value.aggregate.limit = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			copy := request
			aggregateCopy := request.aggregate.clone()
			copy.aggregate = &aggregateCopy
			mutate(&copy)
			if err := copy.Validate(); err == nil {
				t.Fatal("malformed aggregate query validated")
			}
		})
	}
}
