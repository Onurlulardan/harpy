package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func includeRelation(
	t *testing.T, identifier, name string, target api.CatalogReference,
) api.RelationDefinition {
	t.Helper()
	relation, err := api.NewRelationDefinition(
		identifier, name, api.RelationCardinalityToMany,
		includeRelationStorage(t, identifier), target, true, 100,
	)
	if err != nil {
		t.Fatal(err)
	}
	return relation
}

func includeCatalog(
	t *testing.T,
	reference api.CatalogReference,
	tableName string,
	fields []api.FieldDefinition,
	relations []api.RelationDefinition,
	stable api.FieldDefinition,
) api.FieldCatalog {
	t.Helper()
	storage, err := api.NewResourceStorage(fixtureToken(t, tableName))
	if err != nil {
		t.Fatal(err)
	}
	capabilities, err := api.NewCatalogCapabilities(true)
	if err != nil {
		t.Fatal(err)
	}
	stableID, err := stable.ID()
	if err != nil {
		t.Fatal(err)
	}
	defaults := make([]uuid.UUID, len(fields))
	for index, field := range fields {
		defaults[index], err = field.ID()
		if err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, Relations: relations, DefaultSelect: defaults,
		StableID: stableID, SearchFields: []uuid.UUID{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
