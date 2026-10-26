package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRequestAggregateSortRequiresExactReferenceAndDirection(t *testing.T) {
	for _, test := range []struct {
		reference AggregateReference
		name      string
		direction Direction
		valid     bool
	}{
		{AggregateReferenceGroup, "status", DirectionAscending, true},
		{AggregateReferenceMeasure, "row_count", DirectionDescending, true},
		{AggregateReference(99), "status", DirectionAscending, false},
		{AggregateReferenceGroup, "", DirectionAscending, false},
		{AggregateReferenceMeasure, "Upper", DirectionAscending, false},
		{AggregateReferenceGroup, "status", Direction(99), false},
	} {
		value, err := NewRequestAggregateSort(test.reference, test.name, test.direction)
		if (err == nil) != test.valid {
			t.Fatalf("valid=%v sort=%#v err=%v", test.valid, test, err)
		}
		if test.valid && (value.Reference() != test.reference ||
			value.Name() != test.name || value.Direction() != test.direction) {
			t.Fatalf("unexpected aggregate sort: %#v", value)
		}
	}
}

func TestRequestAggregateSortIsRedactedAndCannotMarshal(t *testing.T) {
	sort := mustAggregateSort(t, AggregateReferenceMeasure, "secret_measure")
	for _, value := range []string{fmt.Sprint(sort), fmt.Sprintf("%#v", sort)} {
		if value != queryValueRedaction {
			t.Fatalf("aggregate sort formatting leaked: %q", value)
		}
	}
	if _, err := json.Marshal(sort); err == nil {
		t.Fatal("opaque aggregate sort silently marshaled")
	}
}
