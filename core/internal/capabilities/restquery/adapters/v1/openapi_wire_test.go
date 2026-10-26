package v1query

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestOpenAPIGetFiltersNormalizeToCanonicalPredicates(t *testing.T) {
	decoder := mustDecoder(t)
	tests := []struct {
		raw      string
		field    string
		operator api.Operator
		values   []string
	}{
		{"filter[status]=open", "status", api.OperatorEq, []string{"open"}},
		{"filter[status]=open%2Cclosed", "status", api.OperatorIn,
			[]string{"open", "closed"}},
	}
	for _, test := range tests {
		result, err := decoder.decode(test.raw, DialectOpenAPIGet)
		if err != nil {
			t.Fatalf("decode(%q) error = %v", test.raw, err)
		}
		request, _ := result.Request()
		where, found := request.Where()
		if !found {
			t.Fatal("OpenAPI filter did not produce a canonical predicate")
		}
		requirePredicate(t, where, test.field, test.operator, test.values)
	}
}

func TestOpenAPIAndDynamicDialectsRejectEachOthersFilterKeys(t *testing.T) {
	decoder := mustDecoder(t)
	_, err := decoder.decode("filters[status]=open", DialectOpenAPIGet)
	requireQueryError(t, err, restqueryv1.ErrorQueryInvalid,
		"/filters/status", restqueryv1.IssueMemberUnknown)

	_, err = decoder.decode("filter[status]=open", DialectDynamicList)
	requireQueryError(t, err, restqueryv1.ErrorQueryInvalid,
		"/filter/status", restqueryv1.IssueMemberUnknown)
}

func TestOpenAPIGetPreservesSearchSortAndPage(t *testing.T) {
	result, err := mustDecoder(t).decode(
		"page=3&limit=25&search=needle&sort[field]=status&sort[order]=asc",
		DialectOpenAPIGet,
	)
	if err != nil {
		t.Fatal(err)
	}
	request, _ := result.Request()
	search, found := request.Search()
	page, _ := result.PagePlan()
	pageNumber, _ := page.Page()
	limit, _ := page.Limit()
	sorts := request.Sort()
	if !found || search.Term() != "needle" || pageNumber != 3 || limit != 25 ||
		len(sorts) != 1 || sorts[0].Field() != "status" ||
		sorts[0].Direction() != api.DirectionAscending {
		t.Fatal("OpenAPI search, sort or page normalization drifted")
	}
}
