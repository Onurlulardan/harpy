package resolver

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestResolveAggregateBindsOrdinalsResultTypesAndTieSort(t *testing.T) {
	fixture := newQueryFixture(t, true)
	request := fixtureAggregateRequest(t)
	spec, err := Resolve(request, fixture.set)
	if err != nil || spec.Validate() != nil {
		if queryError, ok := err.(api.QueryError); ok {
			t.Fatalf("aggregate resolution failed: %v", queryError.Issues())
		}
		t.Fatalf("aggregate resolution failed: %v", err)
	}
	aggregate, found, err := spec.Aggregate()
	if err != nil || !found {
		t.Fatalf("aggregate spec missing: %v", err)
	}
	groups, _ := aggregate.Groups()
	measures, _ := aggregate.Measures()
	if len(groups) != 1 || len(measures) != 2 {
		t.Fatal("aggregate declarations were lost")
	}
	assertAggregateResultTypes(t, measures)
	assertAggregateHavingAndSort(t, aggregate)
}

func fixtureAggregateRequest(t *testing.T) api.QueryRequest {
	t.Helper()
	where, _ := api.NewRequestPredicate(
		"amount", api.OperatorGT, []api.Literal{mustTextLiteral(t, "1.20")},
	)
	amount := "amount"
	total, _ := api.NewRequestMeasure(
		"total", api.AggregateFunctionSum, &amount,
	)
	rows, _ := api.NewRequestMeasure("rows", api.AggregateFunctionCount, nil)
	having, _ := api.NewRequestHavingPredicate(
		api.AggregateReferenceMeasure, "total", api.OperatorGT,
		[]api.Literal{mustTextLiteral(t, "2.40")},
	)
	sort, _ := api.NewRequestAggregateSort(
		api.AggregateReferenceMeasure, "total", api.DirectionDescending,
	)
	limit := 20
	aggregate, err := api.NewRequestAggregate(api.RequestAggregateInput{
		GroupBy: []string{"name"}, Measures: []api.RequestMeasure{total, rows},
		Having: &having, Sort: []api.RequestAggregateSort{sort}, Limit: &limit,
	})
	if err != nil {
		t.Fatal(err)
	}
	search, _ := api.NewRequestSearch("Ada")
	request, err := api.NewAggregateQueryRequest(api.AggregateQueryRequestInput{
		Where: &where, Search: &search, Aggregate: aggregate,
	})
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func assertAggregateResultTypes(
	t *testing.T, measures []resolved.AggregateMeasure,
) {
	t.Helper()
	sumType, err := measures[0].ResultType()
	if err != nil {
		t.Fatal(err)
	}
	profile, found, err := sumType.Decimal()
	if err != nil || !found {
		t.Fatal("sum(decimal) did not resolve to decimal")
	}
	precision, _ := profile.Precision()
	scale, _ := profile.Scale()
	countType, _ := measures[1].ResultType()
	integer, integerFound, _ := countType.Integer()
	minimum, _ := integer.Minimum()
	if precision != 22 || scale != 2 || !integerFound || minimum != 0 {
		t.Fatal("aggregate result profiles are not canonical")
	}
}
