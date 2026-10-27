package api

import (
	"encoding/json"
	"testing"
)

func TestAggregateFunctionCanonicalSet(t *testing.T) {
	tests := map[string]AggregateFunction{
		"count": AggregateFunctionCount, "count_distinct": AggregateFunctionCountDistinct,
		"sum": AggregateFunctionSum, "avg": AggregateFunctionAvg,
		"min": AggregateFunctionMin, "max": AggregateFunctionMax,
	}
	for name, want := range tests {
		actual, found := ParseAggregateFunction(name)
		if !found || actual != want || actual.String() != name {
			t.Fatalf("function %q did not round trip", name)
		}
	}
	if value, found := ParseAggregateFunction("median"); found || value != 0 {
		t.Fatal("unknown aggregate function accepted")
	}
}

func TestAggregateEnumsCannotMarshalDirectly(t *testing.T) {
	for _, value := range []any{AggregateFunctionCount, AggregateReferenceGroup} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}
