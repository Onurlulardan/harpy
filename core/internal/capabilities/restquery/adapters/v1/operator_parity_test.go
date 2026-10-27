package v1query

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestV1DirectOperatorCatalogParity(t *testing.T) {
	tests := []struct {
		legacy    string
		canonical api.Operator
	}{
		{"equals", api.OperatorEq},
		{"not_equals", api.OperatorNe},
		{"contains", api.OperatorContains},
		{"starts_with", api.OperatorStartsWith},
		{"ends_with", api.OperatorEndsWith},
		{"greater_than", api.OperatorGT},
		{"less_than", api.OperatorLT},
		{"greater_than_or_equal", api.OperatorGTE},
		{"less_than_or_equal", api.OperatorLTE},
		{"in", api.OperatorIn},
		{"not_in", api.OperatorNotIn},
	}
	decoder := mustDecoder(t)
	for _, test := range tests {
		raw := "filters[status][operator]=" + test.legacy +
			"&filters[status][value]=A"
		result, err := decoder.Decode(raw)
		if err != nil {
			t.Fatalf("Decode(%s) error = %v", test.legacy, err)
		}
		request, _ := result.Request()
		where, _ := request.Where()
		if predicate, found := where.Predicate(); !found ||
			predicate.Operator() != test.canonical {
			t.Fatalf("operator %s did not map to %s", test.legacy, test.canonical)
		}
	}
}

func TestV1IsNotEmptyBuildsExactCanonicalBoolean(t *testing.T) {
	result, err := mustDecoder(t).Decode(
		"filters[status][operator]=is_not_empty&filters[status][value]=ignored",
	)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ := result.Request()
	where, _ := request.Where()
	if where.Kind() != api.RequestFilterAnd || len(where.Children()) != 2 {
		t.Fatalf("is_not_empty filter kind = %v", where.Kind())
	}
	requirePredicate(t, where.Children()[0], "status", api.OperatorIsNotNull, nil)
	requirePredicate(t, where.Children()[1], "status", api.OperatorNe, []string{""})
}
