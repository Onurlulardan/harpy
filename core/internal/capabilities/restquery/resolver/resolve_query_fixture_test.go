package resolver

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	fixtureRootResource   = "10000000-0000-4000-8000-000000000001"
	fixtureRootVersion    = "10000000-0000-4000-8000-000000000002"
	fixtureTargetResource = "20000000-0000-4000-8000-000000000001"
	fixtureTargetVersion  = "20000000-0000-4000-8000-000000000002"
	fixtureRootID         = "30000000-0000-4000-8000-000000000001"
	fixtureRootName       = "30000000-0000-4000-8000-000000000002"
	fixtureRootAmount     = "30000000-0000-4000-8000-000000000003"
	fixtureTargetID       = "40000000-0000-4000-8000-000000000001"
	fixtureTargetName     = "40000000-0000-4000-8000-000000000002"
)

type queryFixture struct {
	set  api.CatalogSet
	root api.FieldCatalog
}

func newQueryFixture(t *testing.T, exactCount bool) queryFixture {
	t.Helper()
	rootReference := fixtureReference(
		t, "forms", fixtureRootResource, fixtureRootVersion, "1",
	)
	targetReference := fixtureReference(
		t, "crm", fixtureTargetResource, fixtureTargetVersion, "2",
	)
	children, err := api.NewRelationDefinition(
		"50000000-0000-4000-8000-000000000001", "children",
		api.RelationCardinalityToMany,
		fixtureRelationStorage(t, "50000000-0000-4000-8000-000000000001"),
		targetReference, true, 100,
	)
	if err != nil {
		t.Fatal(err)
	}
	parent, err := api.NewRelationDefinition(
		"50000000-0000-4000-8000-000000000002", "parent",
		api.RelationCardinalityToOne,
		fixtureRelationStorage(t, "50000000-0000-4000-8000-000000000002"),
		rootReference, true, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	rootFields := []api.FieldDefinition{
		fixtureUUIDField(t, fixtureRootID, "id", "fld_root_id"),
		fixtureTextField(t, fixtureRootName, "name", "root_name"),
		fixtureDecimalField(t, fixtureRootAmount, "amount", "fld_root_amount"),
	}
	targetFields := []api.FieldDefinition{
		fixtureUUIDField(t, fixtureTargetID, "id", "fld_target_id"),
		fixtureTextField(t, fixtureTargetName, "name", "target_name"),
	}
	root := fixtureCatalog(
		t, rootReference, "frm_root_records", rootFields,
		[]api.RelationDefinition{children}, fixtureRootID, fixtureRootName, exactCount,
	)
	target := fixtureCatalog(
		t, targetReference, "frm_target_records", targetFields,
		[]api.RelationDefinition{parent}, fixtureTargetID, fixtureTargetName, true,
	)
	set, err := api.NewCatalogSet(rootReference, []api.FieldCatalog{root, target})
	if err != nil {
		t.Fatal(err)
	}
	return queryFixture{set: set, root: root}
}

func fixtureRelationStorage(t *testing.T, relationID string) api.RelationStorage {
	t.Helper()
	name, err := physicalid.Name("rel_", uuid.MustParse(relationID))
	if err != nil {
		t.Fatal(err)
	}
	table, err := sqlidentifier.MintDynamic(name)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := api.NewRelationStorage(relationID, table)
	if err != nil {
		t.Fatal(err)
	}
	return storage
}

func fixtureReference(
	t *testing.T, owner, resource, version, digit string,
) api.CatalogReference {
	t.Helper()
	hash, err := api.NewCatalogHash(strings.Repeat(digit, 64))
	if err != nil {
		t.Fatal(err)
	}
	reference, err := api.NewCatalogReference(owner, resource, version, hash)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func fixtureCatalog(
	t *testing.T,
	reference api.CatalogReference,
	tableName string,
	fields []api.FieldDefinition,
	relations []api.RelationDefinition,
	stableID string,
	searchID string,
	exactCount bool,
) api.FieldCatalog {
	t.Helper()
	table, _ := sqlidentifier.MintDynamic(tableName)
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(exactCount)
	stable := uuid.MustParse(stableID)
	search := uuid.MustParse(searchID)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, Relations: relations,
		DefaultSelect: []uuid.UUID{stable, search}, StableID: stable,
		SearchFields: []uuid.UUID{search},
	})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
