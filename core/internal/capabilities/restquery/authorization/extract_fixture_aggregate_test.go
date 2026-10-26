package authorization

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

func authorizationAggregateSpec(
	t *testing.T,
	fixture authorizationQueryFixture,
) resolved.QuerySpec {
	t.Helper()
	amount := "amount"
	sum, _ := api.NewRequestMeasure(
		"total", api.AggregateFunctionSum, &amount,
	)
	count, _ := api.NewRequestMeasure(
		"rows", api.AggregateFunctionCount, nil,
	)
	groupHaving := authHaving(
		t, api.AggregateReferenceGroup, "name", api.OperatorEq, "Ada",
	)
	countHaving := authHaving(
		t, api.AggregateReferenceMeasure, "rows", api.OperatorGT, "0",
	)
	having, err := api.NewRequestHavingAnd(
		[]api.RequestHaving{groupHaving, countHaving},
	)
	if err != nil {
		t.Fatal(err)
	}
	groupSort, _ := api.NewRequestAggregateSort(
		api.AggregateReferenceGroup, "name", api.DirectionAscending,
	)
	countSort, _ := api.NewRequestAggregateSort(
		api.AggregateReferenceMeasure, "rows", api.DirectionDescending,
	)
	aggregate, err := api.NewRequestAggregate(api.RequestAggregateInput{
		GroupBy:  []string{"name"},
		Measures: []api.RequestMeasure{sum, count}, Having: &having,
		Sort: []api.RequestAggregateSort{groupSort, countSort},
	})
	if err != nil {
		t.Fatal(err)
	}
	where := authPredicate(t, "amount", api.OperatorGT, "1.20")
	search, _ := api.NewRequestSearch("Ada")
	request, err := api.NewAggregateQueryRequest(api.AggregateQueryRequestInput{
		Where: &where, Search: &search, Aggregate: aggregate,
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

func authHaving(
	t *testing.T,
	reference api.AggregateReference,
	name string,
	operator api.Operator,
	value string,
) api.RequestHaving {
	t.Helper()
	literal, _ := api.NewTextLiteral(value)
	having, err := api.NewRequestHavingPredicate(
		reference, name, operator, []api.Literal{literal},
	)
	if err != nil {
		t.Fatal(err)
	}
	return having
}
