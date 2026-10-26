package v1query

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestV1FilterSourceShapesBecomeCanonicalPredicates(t *testing.T) {
	// Characterizes V1 ParseQueryParams and ApplyFilters without carrying SQL.
	tests := []struct {
		name     string
		raw      string
		field    string
		operator api.Operator
		values   []string
	}{
		{"old scalar", "filters[status]=OPEN", "status", api.OperatorEq, []string{"OPEN"}},
		{
			"new operator", "filters[status][operator]=contains&filters[status][value]=pen",
			"status", api.OperatorContains, []string{"pen"},
		},
		{
			"list operator", "filters[status][operator]=in&filters[status][value]=A%2C+B",
			"status", api.OperatorIn, []string{"A", "B"},
		},
		{
			"array notation", "filters[tags][]=red&filters[tags][]=&" +
				"filters[tags][]=blue%2Cgreen&filters[tags][]=%20%20",
			"tags", api.OperatorContainsAny, []string{"red", "blue", "green"},
		},
		{
			"explicit array operator",
			"filters[tags][operator]=array_contains&filters[tags][value]=red%2Cblue",
			"tags", api.OperatorContainsAny, []string{"red", "blue"},
		},
	}
	decoder := mustDecoder(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := decoder.Decode(test.raw)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			request, _ := result.Request()
			where, found := request.Where()
			if !found {
				t.Fatal("canonical predicate is missing")
			}
			requirePredicate(t, where, test.field, test.operator, test.values)
		})
	}
}

func TestV1SpecialOperatorsPreserveSourceSemantics(t *testing.T) {
	decoder := mustDecoder(t)
	result, err := decoder.Decode(
		"filters[status][operator]=is_empty&filters[status][value]=ignored",
	)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ := result.Request()
	where, _ := request.Where()
	if where.Kind() != api.RequestFilterOr || len(where.Children()) != 2 {
		t.Fatalf("is_empty filter kind = %v, children = %d", where.Kind(), len(where.Children()))
	}
	requirePredicate(t, where.Children()[0], "status", api.OperatorIsNull, nil)
	requirePredicate(t, where.Children()[1], "status", api.OperatorEq, []string{""})

	result, err = decoder.Decode(
		"filters[status][operator]=not_contains&filters[status][value]=closed",
	)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ = result.Request()
	where, _ = request.Where()
	if where.Kind() != api.RequestFilterNot || len(where.Children()) != 1 {
		t.Fatalf("not_contains filter kind = %v", where.Kind())
	}
	requirePredicate(
		t, where.Children()[0], "status", api.OperatorContains, []string{"closed"},
	)
}

func TestV1GlobalFilterOperatorBuildsOneBooleanRoot(t *testing.T) {
	result, err := mustDecoder(t).Decode(
		"filter_operator=or&filters[status]=OPEN&filters[priority]=HIGH",
	)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ := result.Request()
	where, _ := request.Where()
	if where.Kind() != api.RequestFilterOr || len(where.Children()) != 2 {
		t.Fatalf("where kind = %v, children = %d", where.Kind(), len(where.Children()))
	}
}

func TestV1IncompleteFilterRemainsAbsent(t *testing.T) {
	result, err := mustDecoder(t).Decode("filters[status][operator]=equals")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ := result.Request()
	if _, found := request.Where(); found {
		t.Fatal("incomplete V1 filter unexpectedly became a predicate")
	}
}

func TestV1FilterUnknownsFailClosed(t *testing.T) {
	decoder := mustDecoder(t)
	tests := []struct {
		raw  string
		path string
		code restqueryv1.IssueCode
	}{
		{"filters[missing]=x", "/filters/missing", restqueryv1.IssueFieldUnknown},
		{
			"filters[status][operator]=mystery&filters[status][value]=x",
			"/filters/status/operator", restqueryv1.IssueOperatorUnknown,
		},
		{"filter_operator=XOR", "/filter_operator", restqueryv1.IssueOperatorUnknown},
		{
			"filters[status]=x&filters[status][operator]=equals",
			"/filters/status", restqueryv1.IssueMemberDuplicate,
		},
		{
			"filters[tags]=x", "/filters/tags",
			restqueryv1.IssueOperatorNotAllowed,
		},
	}
	for _, test := range tests {
		_, err := decoder.Decode(test.raw)
		requireQueryError(t, err, restqueryv1.ErrorQueryInvalid, test.path, test.code)
	}
}
