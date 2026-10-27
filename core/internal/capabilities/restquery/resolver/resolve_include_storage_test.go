package resolver

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveIncludeBindsExactTargetAndRelationStorage(t *testing.T) {
	fixture := newQueryFixture(t, true)
	include, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Select: []string{"name"},
	})
	if err != nil {
		t.Fatal(err)
	}
	page, _ := api.NewRequestPage(10, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Include: []api.RequestInclude{include},
		Page: page, Count: api.CountModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := Resolve(request, fixture.set)
	if err != nil {
		t.Fatal(err)
	}
	record, found, err := spec.Record()
	if err != nil || !found {
		t.Fatal("resolved record query missing")
	}
	includes, err := record.Includes()
	if err != nil || len(includes) != 1 {
		t.Fatal("resolved include missing")
	}
	actualTarget, targetErr := includes[0].TargetStorage()
	actualRelation, relationErr := includes[0].RelationStorage()
	relation, found, relationLookupErr := fixture.root.RelationByPublicName("children")
	if relationLookupErr != nil || !found {
		t.Fatal("fixture relation missing")
	}
	relationStorage, _ := relation.Storage()
	targetCatalog, found, _ := fixture.set.Catalog(mustRelationTargetReference(t, fixture))
	if !found {
		t.Fatal("target catalog missing")
	}
	targetStorage, _ := targetCatalog.Storage()
	if targetErr != nil || relationErr != nil ||
		actualTarget != targetStorage || actualRelation != relationStorage {
		t.Fatal("include did not bind exact target and relation storage")
	}
}

func mustRelationTargetReference(t *testing.T, fixture queryFixture) api.CatalogReference {
	t.Helper()
	relation, found, err := fixture.root.RelationByPublicName("children")
	if err != nil || !found {
		t.Fatal("fixture relation missing")
	}
	target, err := relation.Target()
	if err != nil {
		t.Fatal(err)
	}
	return target
}
