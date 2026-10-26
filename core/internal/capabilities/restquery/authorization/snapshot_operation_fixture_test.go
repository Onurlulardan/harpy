package authorization

import (
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

func snapshotRecordSpec(t *testing.T, ownerCount int) resolved.QuerySpec {
	t.Helper()
	if ownerCount < 1 || ownerCount > 4 {
		t.Fatal("invalid snapshot fixture owner count")
	}
	references := []api.CatalogReference{
		snapshotReference(t, "forms", 1),
		snapshotReference(t, "crm", 2),
		snapshotReference(t, "documents", 3),
		snapshotReference(t, "workflow", 4),
		snapshotReference(t, "iam", 5),
	}
	rootRelations := []api.RelationDefinition{
		snapshotRelation(t, 11, "children", references[1]),
		snapshotRelation(t, 12, "documents", references[2]),
	}
	crmRelations := []api.RelationDefinition{
		snapshotRelation(t, 13, "workflow", references[3]),
	}
	catalogs, err := api.NewCatalogSet(references[0], []api.FieldCatalog{
		snapshotCatalog(t, references[0], "snapshot_forms", rootRelations),
		snapshotCatalog(t, references[1], "snapshot_crm", crmRelations),
		snapshotCatalog(t, references[2], "snapshot_documents", nil),
		snapshotCatalog(t, references[3], "snapshot_workflow", nil),
		snapshotCatalog(t, references[4], "snapshot_unused_iam", nil),
	})
	if err != nil {
		t.Fatal(err)
	}
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Include: snapshotIncludes(t, ownerCount),
		Page: snapshotPage(t), Count: api.CountModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func snapshotIncludes(t *testing.T, ownerCount int) []api.RequestInclude {
	t.Helper()
	if ownerCount == 1 {
		return nil
	}
	var nested []api.RequestInclude
	if ownerCount >= 3 {
		nested = []api.RequestInclude{snapshotInclude(t, "workflow", nil)}
	}
	result := []api.RequestInclude{snapshotInclude(t, "children", nested)}
	if ownerCount == 4 {
		result = append(result, snapshotInclude(t, "documents", nil))
	}
	return result
}

func snapshotInclude(
	t *testing.T,
	relation string,
	children []api.RequestInclude,
) api.RequestInclude {
	t.Helper()
	include, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: relation, Select: []string{"name"}, Include: children,
	})
	if err != nil {
		t.Fatal(err)
	}
	return include
}

func snapshotReference(
	t *testing.T,
	owner string,
	index int,
) api.CatalogReference {
	t.Helper()
	resource := snapshotUUID(index)
	version := snapshotUUID(index + 100)
	return authReference(t, owner, resource, version, fmt.Sprintf("%x", index))
}

func snapshotRelation(
	t *testing.T,
	index int,
	name string,
	target api.CatalogReference,
) api.RelationDefinition {
	t.Helper()
	relation, err := api.NewRelationDefinition(
		snapshotUUID(index), name, api.RelationCardinalityToMany,
		authRelationStorage(t, snapshotUUID(index)), target, true, 100,
	)
	if err != nil {
		t.Fatal(err)
	}
	return relation
}

func snapshotCatalog(
	t *testing.T,
	reference api.CatalogReference,
	table string,
	relations []api.RelationDefinition,
) api.FieldCatalog {
	t.Helper()
	fields := []api.FieldDefinition{
		authUUIDField(t, authTargetID, "id", table+"_id"),
		authTextField(t, authTargetName, "name", table+"_name"),
	}
	return authCatalog(
		t, reference, table, fields, relations, authTargetID, authTargetName,
	)
}

func snapshotUUID(index int) string {
	return fmt.Sprintf("%08x-0000-4000-8000-%012x", index, index)
}

func snapshotPage(t *testing.T) api.RequestPage {
	t.Helper()
	page, err := api.NewRequestPage(25, nil)
	if err != nil {
		t.Fatal(err)
	}
	return page
}
