package authorization

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestMandatoryFilterDepthAndPredicateBounds(t *testing.T) {
	filter := mandatoryLeaf(t)
	var err error
	for depth := 0; depth < maximumMandatoryDepth; depth++ {
		filter, err = NewMandatoryNot(filter)
		if err != nil {
			t.Fatalf("depth %d: %v", depth+1, err)
		}
	}
	if _, err := NewMandatoryNot(filter); err == nil {
		t.Fatal("mandatory depth overflow was accepted")
	}

	children := make([]MandatoryFilter, maximumMandatoryPredicates)
	for index := range children {
		children[index] = mandatoryLeaf(t)
	}
	if _, err := NewMandatoryAnd(children); err != nil {
		t.Fatalf("exact mandatory predicate bound: %v", err)
	}
	children = append(children, mandatoryLeaf(t))
	if _, err := NewMandatoryAnd(children); err == nil {
		t.Fatal("mandatory predicate overflow was accepted")
	}
}

func TestMandatoryFilterLiteralAndBooleanShapeBounds(t *testing.T) {
	values := make([]resolved.Value, maximumMandatoryLiterals)
	for index := range values {
		values[index] = mandatoryScalar(t, api.LogicalTypeText, "value")
	}
	predicate, err := NewMandatoryPredicate(
		testFieldID, api.LogicalTypeText, api.OperatorIn, values,
	)
	if err != nil {
		t.Fatalf("exact mandatory literal bound: %v", err)
	}
	if _, err := NewMandatoryPredicateFilter(predicate); err != nil {
		t.Fatal(err)
	}
	values = append(values, mandatoryScalar(t, api.LogicalTypeText, "overflow"))
	if _, err := NewMandatoryPredicate(
		testFieldID, api.LogicalTypeText, api.OperatorIn, values,
	); err == nil {
		t.Fatal("mandatory literal overflow was accepted")
	}
	if _, err := NewMandatoryAnd([]MandatoryFilter{mandatoryLeaf(t)}); err == nil {
		t.Fatal("single-child mandatory AND was accepted")
	}
}
