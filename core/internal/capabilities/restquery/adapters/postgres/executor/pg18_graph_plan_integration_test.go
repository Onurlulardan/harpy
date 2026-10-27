//go:build integration

package executor

import (
	"crypto/sha256"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/google/uuid"
)

func pg18GraphPlan(
	t *testing.T, tenant string, count api.CountMode, pageSize int,
) graphPlanFixture {
	t.Helper()
	fixture := newGraphPlanFixture(t, count, pageSize)
	spec, err := fixture.plan.Spec()
	if err != nil {
		t.Fatal(err)
	}
	fixture.plan = pg18MintGraphPlan(t, tenant, spec)
	return fixture
}

func pg18GraphInjectionPlan(
	t *testing.T, tenant string, payload string,
) graphPlanFixture {
	t.Helper()
	catalogs := newGraphCatalogFixture(t)
	labelSort, _ := api.NewRequestSort("label", api.DirectionAscending)
	childLimit, childrenLimit := 1, 1
	child, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "items", Select: []string{"label"},
		Sort: []api.RequestSort{labelSort}, Limit: &childLimit,
	})
	if err != nil {
		t.Fatal(err)
	}
	where := pg18GraphLabelFilter(t, payload)
	children, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Select: []string{"label"}, Where: where,
		Sort: []api.RequestSort{labelSort}, Limit: &childrenLimit,
		Include: []api.RequestInclude{child},
	})
	if err != nil {
		t.Fatal(err)
	}
	owner, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "owner", Select: []string{"label"},
		Sort: []api.RequestSort{labelSort},
	})
	if err != nil {
		t.Fatal(err)
	}
	page, _ := api.NewRequestPage(10, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"label"}, Sort: []api.RequestSort{labelSort}, Page: page,
		Count: api.CountModeExact, Include: []api.RequestInclude{children, owner},
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, catalogs.set)
	if err != nil {
		t.Fatal(err)
	}
	return graphPlanFixture{plan: pg18MintGraphPlan(t, tenant, spec), catalogs: catalogs}
}

func pg18MintGraphPlan(
	t *testing.T, tenant string, spec resolved.QuerySpec,
) resolved.AuthorizedPlan {
	t.Helper()
	plan, err := resolved.MintAuthorizedPlan(
		spec, uuid.MustParse(tenant), uuid.MustParse(pg18Actor),
		sha256.Sum256([]byte("pg18-graph-authorization")),
		graphVisibilities(t, spec),
	)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func pg18GraphLabelFilter(t *testing.T, value string) *api.RequestFilter {
	t.Helper()
	literal, err := api.NewTextLiteral(value)
	if err != nil {
		t.Fatal(err)
	}
	filter, err := api.NewRequestPredicate(
		"label", api.OperatorEq, []api.Literal{literal},
	)
	if err != nil {
		t.Fatal(err)
	}
	return &filter
}
