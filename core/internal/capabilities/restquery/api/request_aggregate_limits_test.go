package api

import "testing"

func TestRequestAggregateEnforcesGroupMeasureAndSortLimits(t *testing.T) {
	base := RequestAggregateInput{}
	base.GroupBy = []string{"status", "region", "day"}
	for index := 0; index < maximumAggregateMeasures; index++ {
		base.Measures = append(base.Measures, mustRequestMeasure(
			t, requestMeasureName(index), AggregateFunctionCount, nil,
		))
	}
	base.Sort = []RequestAggregateSort{
		mustAggregateSort(t, AggregateReferenceGroup, "status"),
		mustAggregateSort(t, AggregateReferenceGroup, "region"),
		mustAggregateSort(t, AggregateReferenceGroup, "day"),
	}
	if _, err := NewRequestAggregate(base); err != nil {
		t.Fatalf("hard maxima rejected: %v", err)
	}
	groups := base
	groups.GroupBy = append(groups.GroupBy, "fourth")
	if _, err := NewRequestAggregate(groups); err == nil {
		t.Fatal("four group fields accepted")
	}
	measures := base
	measures.Measures = append(measures.Measures,
		mustRequestMeasure(t, "ninth", AggregateFunctionCount, nil))
	if _, err := NewRequestAggregate(measures); err == nil {
		t.Fatal("nine measures accepted")
	}
	sorts := base
	sorts.Sort = append(sorts.Sort,
		mustAggregateSort(t, AggregateReferenceMeasure, "m_0"))
	if _, err := NewRequestAggregate(sorts); err == nil {
		t.Fatal("four aggregate sorts accepted")
	}
}

func TestRequestAggregateEnforcesUniqueDeclarationsAndReferences(t *testing.T) {
	measure := mustRequestMeasure(t, "row_count", AggregateFunctionCount, nil)
	for _, input := range []RequestAggregateInput{
		{GroupBy: []string{"status", "status"}},
		{Measures: []RequestMeasure{measure, measure}},
		{Measures: []RequestMeasure{measure}, Having: havingPointer(t,
			AggregateReferenceGroup, "unknown")},
		{GroupBy: []string{"status"}, Measures: []RequestMeasure{measure},
			Sort: []RequestAggregateSort{mustAggregateSort(t, AggregateReferenceMeasure, "unknown")}},
		{GroupBy: []string{"status"}, Sort: []RequestAggregateSort{
			mustAggregateSort(t, AggregateReferenceGroup, "status"),
			mustAggregateSort(t, AggregateReferenceGroup, "status"),
		}},
	} {
		if _, err := NewRequestAggregate(input); err == nil {
			t.Fatalf("invalid declarations accepted: %#v", input)
		}
	}
}

func TestRequestAggregateLimitDefaultAndBounds(t *testing.T) {
	measure := mustRequestMeasure(t, "row_count", AggregateFunctionCount, nil)
	base := RequestAggregateInput{Measures: []RequestMeasure{measure}}
	if value, err := NewRequestAggregate(base); err != nil ||
		value.Limit() != defaultAggregateLimit {
		t.Fatalf("default limit failed: %d, %v", value.Limit(), err)
	}
	for _, limit := range []int{1, maximumAggregateLimit} {
		input := base
		input.Limit = &limit
		value, err := NewRequestAggregate(input)
		if err != nil || value.Limit() != limit {
			t.Fatalf("limit %d rejected: %v", limit, err)
		}
	}
	for _, limit := range []int{0, maximumAggregateLimit + 1} {
		input := base
		input.Limit = &limit
		if _, err := NewRequestAggregate(input); err == nil {
			t.Fatalf("limit %d accepted", limit)
		}
	}
}

func TestRequestAggregateValidateRejectsZeroAndNilCollections(t *testing.T) {
	if err := (RequestAggregate{}).Validate(); err == nil {
		t.Fatal("zero aggregate validated")
	}
	if _, err := NewRequestAggregate(RequestAggregateInput{}); err == nil {
		t.Fatal("empty aggregate declaration was accepted")
	}
	aggregate, _ := NewRequestAggregate(RequestAggregateInput{Measures: []RequestMeasure{
		mustRequestMeasure(t, "row_count", AggregateFunctionCount, nil),
	}})
	aggregate.sorts = nil
	if err := aggregate.Validate(); err == nil {
		t.Fatal("nil canonical collection validated")
	}
}

func requestMeasureName(index int) string { return "m_" + string(rune('0'+index)) }

func havingPointer(t *testing.T, reference AggregateReference, name string) *RequestHaving {
	t.Helper()
	value := mustHavingPredicate(t, reference, name, OperatorIsNull, nil)
	return &value
}
