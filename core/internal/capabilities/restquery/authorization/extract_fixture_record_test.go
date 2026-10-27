package authorization

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

func authorizationRecordSpec(
	t *testing.T,
	fixture authorizationQueryFixture,
) resolved.QuerySpec {
	t.Helper()
	where := authPredicate(t, "amount", api.OperatorGT, "1.20")
	includeWhere := authPredicate(t, "name", api.OperatorContains, "Ada")
	includeSort, _ := api.NewRequestSort("name", api.DirectionDescending)
	include, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Select: []string{"name"},
		Where: &includeWhere, Sort: []api.RequestSort{includeSort},
	})
	if err != nil {
		t.Fatal(err)
	}
	rootSort, _ := api.NewRequestSort("name", api.DirectionAscending)
	page, _ := api.NewRequestPage(25, nil)
	search, _ := api.NewRequestSearch("Ada")
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"amount"}, Where: &where,
		Sort: []api.RequestSort{rootSort}, Page: page,
		Search: &search, Count: api.CountModeExact,
		Include: []api.RequestInclude{include},
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, fixture.catalogs)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func authPredicate(
	t *testing.T,
	field string,
	operator api.Operator,
	value string,
) api.RequestFilter {
	t.Helper()
	literal, err := api.NewTextLiteral(value)
	if err != nil {
		t.Fatal(err)
	}
	filter, err := api.NewRequestPredicate(
		field, operator, []api.Literal{literal},
	)
	if err != nil {
		t.Fatal(err)
	}
	return filter
}
