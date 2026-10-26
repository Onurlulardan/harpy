package executor

import (
	"crypto/sha256"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

type graphPlanFixture struct {
	plan     resolved.AuthorizedPlan
	catalogs graphCatalogFixture
}

func newGraphPlanFixture(
	t *testing.T,
	count api.CountMode,
	pageSize int,
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
	children, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Select: []string{"label"},
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
	page, _ := api.NewRequestPage(pageSize, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"label"}, Sort: []api.RequestSort{labelSort},
		Page: page, Count: count, Include: []api.RequestInclude{children, owner},
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, catalogs.set)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := resolved.MintAuthorizedPlan(
		spec, fixtureTenant, fixtureActor,
		sha256.Sum256([]byte("graph-authorization")),
		graphVisibilities(t, spec),
	)
	if err != nil {
		t.Fatal(err)
	}
	return graphPlanFixture{plan: plan, catalogs: catalogs}
}

func graphVisibilities(
	t *testing.T,
	spec resolved.QuerySpec,
) []resolved.ResourceVisibility {
	t.Helper()
	reference, referenceErr := spec.Reference()
	rootPath, pathErr := spec.RootPath()
	if referenceErr != nil || pathErr != nil {
		t.Fatal("root visibility identity is invalid")
	}
	result := []resolved.ResourceVisibility{
		graphVisibility(t, reference, rootPath),
	}
	record, found, err := spec.Record()
	if err != nil || !found {
		t.Fatal("record query is missing")
	}
	includes, err := record.Includes()
	if err != nil {
		t.Fatal(err)
	}
	return append(result, graphIncludeVisibilities(t, includes)...)
}

func graphIncludeVisibilities(
	t *testing.T,
	includes []resolved.Include,
) []resolved.ResourceVisibility {
	t.Helper()
	result := make([]resolved.ResourceVisibility, 0, len(includes))
	for _, include := range includes {
		reference, referenceErr := include.Target()
		path, pathErr := include.Path()
		if referenceErr != nil || pathErr != nil {
			t.Fatal("include visibility identity is invalid")
		}
		result = append(result, graphVisibility(t, reference, path))
		children, err := include.Children()
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, graphIncludeVisibilities(t, children)...)
	}
	return result
}

func graphVisibility(
	t *testing.T,
	reference api.CatalogReference,
	path resolved.RequestPath,
) resolved.ResourceVisibility {
	t.Helper()
	value, _ := path.Value()
	visibility, err := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("graph-visibility:"+value)), reference, path,
	)
	if err != nil {
		t.Fatal(err)
	}
	return visibility
}
