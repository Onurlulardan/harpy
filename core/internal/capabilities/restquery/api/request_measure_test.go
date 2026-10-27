package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestRequestMeasureEnforcesNameAndFieldArity(t *testing.T) {
	field := "total"
	count, err := NewRequestMeasure("row_count", AggregateFunctionCount, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := count.Field(); found {
		t.Fatal("count gained a field")
	}
	if _, err := NewRequestMeasure("row_count", AggregateFunctionCount, &field); err == nil {
		t.Fatal("count accepted a field")
	}
	for _, function := range []AggregateFunction{
		AggregateFunctionCountDistinct, AggregateFunctionSum, AggregateFunctionAvg,
		AggregateFunctionMin, AggregateFunctionMax,
	} {
		if _, err := NewRequestMeasure("value", function, nil); err == nil {
			t.Fatalf("%s accepted an absent field", function)
		}
		if _, err := NewRequestMeasure("value", function, &field); err != nil {
			t.Fatalf("%s rejected a field: %v", function, err)
		}
	}
}

func TestRequestMeasureNameGrammar(t *testing.T) {
	field := "field"
	invalid := []string{"", "1name", "Upper", "bad-name", strings.Repeat("a", 64)}
	for _, name := range invalid {
		if _, err := NewRequestMeasure(name, AggregateFunctionSum, &field); err == nil {
			t.Fatalf("invalid measure name %q accepted", name)
		}
	}
	if _, err := NewRequestMeasure("a_1", AggregateFunctionSum, &field); err != nil {
		t.Fatalf("valid measure name rejected: %v", err)
	}
}

func TestRequestMeasureIsRedactedAndCannotMarshal(t *testing.T) {
	field := "secret_field"
	measure, _ := NewRequestMeasure("secret_measure", AggregateFunctionSum, &field)
	for _, value := range []string{fmt.Sprint(measure), fmt.Sprintf("%#v", measure)} {
		if value != queryValueRedaction {
			t.Fatalf("measure formatting leaked: %q", value)
		}
	}
	if _, err := json.Marshal(measure); err == nil {
		t.Fatal("opaque measure silently marshaled")
	}
}
