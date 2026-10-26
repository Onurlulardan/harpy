package resolver

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveRejectsRelationCycleAtCanonicalPath(t *testing.T) {
	fixture := newQueryFixture(t, true)
	parent, err := api.NewRequestInclude(api.RequestIncludeInput{Relation: "parent"})
	if err != nil {
		t.Fatal(err)
	}
	children, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Include: []api.RequestInclude{parent},
	})
	if err != nil {
		t.Fatal(err)
	}
	page, _ := api.NewRequestPage(10, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Page: page, Count: api.CountModeNone,
		Include: []api.RequestInclude{children},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertResolutionError(
		t, request, fixture.set, restqueryv1.ErrorQueryInvalid,
		"/include/0/include/0/relation", restqueryv1.IssueRelationCycle,
	)
}

func TestResolveRejectsDisallowedAggregateFunction(t *testing.T) {
	fixture := newQueryFixture(t, true)
	name := "name"
	measure, _ := api.NewRequestMeasure("total", api.AggregateFunctionSum, &name)
	aggregate, _ := api.NewRequestAggregate(api.RequestAggregateInput{
		Measures: []api.RequestMeasure{measure},
	})
	request, _ := api.NewAggregateQueryRequest(api.AggregateQueryRequestInput{
		Aggregate: aggregate,
	})
	assertResolutionError(
		t, request, fixture.set, restqueryv1.ErrorQueryInvalid,
		"/aggregate/measures/0/function", restqueryv1.IssueFunctionNotAllowed,
	)
}
