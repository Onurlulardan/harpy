package authorization

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func mandatoryScalar(
	t *testing.T, kind api.LogicalType, value string,
) resolved.Value {
	t.Helper()
	result, err := resolved.NewScalarValue(kind, value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mandatoryTextPredicate(t *testing.T) MandatoryPredicate {
	t.Helper()
	predicate, err := NewMandatoryPredicate(
		testFieldID,
		api.LogicalTypeText,
		api.OperatorEq,
		[]resolved.Value{mandatoryScalar(t, api.LogicalTypeText, "visible")},
	)
	if err != nil {
		t.Fatal(err)
	}
	return predicate
}

func mandatoryLeaf(t *testing.T) MandatoryFilter {
	t.Helper()
	filter, err := NewMandatoryPredicateFilter(mandatoryTextPredicate(t))
	if err != nil {
		t.Fatal(err)
	}
	return filter
}
