package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRequestAggregateBuildsImmutableCanonicalValue(t *testing.T) {
	count := mustRequestMeasure(t, "row_count", AggregateFunctionCount, nil)
	totalField := "total"
	revenue := mustRequestMeasure(t, "revenue", AggregateFunctionSum, &totalField)
	literal := mustLiteral(t, `"1000.00"`)
	having := mustHavingPredicate(t, AggregateReferenceMeasure, "revenue", OperatorGTE,
		[]Literal{literal})
	sort := mustAggregateSort(t, AggregateReferenceGroup, "status")
	input := RequestAggregateInput{
		GroupBy: []string{"status"}, Measures: []RequestMeasure{count, revenue},
		Having: &having, Sort: []RequestAggregateSort{sort},
	}
	aggregate, err := NewRequestAggregate(input)
	if err != nil {
		t.Fatal(err)
	}
	input.GroupBy[0] = "changed"
	input.Measures[0] = RequestMeasure{}
	input.Sort[0] = RequestAggregateSort{}
	having = RequestHaving{}
	if aggregate.Limit() != defaultAggregateLimit || aggregate.GroupBy()[0] != "status" {
		t.Fatalf("unexpected aggregate: %#v", aggregate)
	}
	groups := aggregate.GroupBy()
	groups[0] = "changed"
	measures := aggregate.Measures()
	measures[0] = RequestMeasure{}
	sorts := aggregate.Sort()
	sorts[0] = RequestAggregateSort{}
	returnedHaving, found := aggregate.Having()
	if !found {
		t.Fatal("aggregate lost HAVING")
	}
	returnedHaving.children = nil
	if aggregate.GroupBy()[0] != "status" || aggregate.Measures()[0].Name() != "row_count" ||
		aggregate.Sort()[0].Name() != "status" || !aggregate.having.valid() {
		t.Fatal("aggregate exposed mutable storage")
	}
}

func TestRequestAggregateIsRedactedAndCannotMarshal(t *testing.T) {
	input := RequestAggregateInput{
		GroupBy:  []string{"secret_group"},
		Measures: []RequestMeasure{mustRequestMeasure(t, "secret_count", AggregateFunctionCount, nil)},
	}
	aggregate, err := NewRequestAggregate(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{
		fmt.Sprint(aggregate), fmt.Sprintf("%#v", aggregate),
		fmt.Sprint(input), fmt.Sprintf("%#v", input),
	} {
		if value != queryValueRedaction {
			t.Fatalf("aggregate formatting leaked: %q", value)
		}
	}
	for _, value := range []any{aggregate, input} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}

func mustRequestMeasure(
	t *testing.T, name string, function AggregateFunction, field *string,
) RequestMeasure {
	t.Helper()
	measure, err := NewRequestMeasure(name, function, field)
	if err != nil {
		t.Fatal(err)
	}
	return measure
}

func mustAggregateSort(
	t *testing.T, reference AggregateReference, name string,
) RequestAggregateSort {
	t.Helper()
	sort, err := NewRequestAggregateSort(reference, name, DirectionAscending)
	if err != nil {
		t.Fatal(err)
	}
	return sort
}
