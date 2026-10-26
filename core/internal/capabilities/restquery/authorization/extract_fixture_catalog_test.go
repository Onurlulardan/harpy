package authorization

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	authRootResource   = "11000000-0000-4000-8000-000000000001"
	authRootVersion    = "11000000-0000-4000-8000-000000000002"
	authTargetResource = "12000000-0000-4000-8000-000000000001"
	authTargetVersion  = "12000000-0000-4000-8000-000000000002"
	authRootID         = "13000000-0000-4000-8000-000000000001"
	authRootName       = "13000000-0000-4000-8000-000000000002"
	authRootAmount     = "13000000-0000-4000-8000-000000000003"
	authTargetID       = "14000000-0000-4000-8000-000000000001"
	authTargetName     = "14000000-0000-4000-8000-000000000002"
	authChildren       = "15000000-0000-4000-8000-000000000001"
)

type authorizationQueryFixture struct {
	catalogs api.CatalogSet
}

func newAuthorizationQueryFixture(t *testing.T) authorizationQueryFixture {
	t.Helper()
	rootReference := authReference(
		t, "forms", authRootResource, authRootVersion, "a",
	)
	targetReference := authReference(
		t, "crm", authTargetResource, authTargetVersion, "b",
	)
	children, err := api.NewRelationDefinition(
		authChildren, "children", api.RelationCardinalityToMany,
		authRelationStorage(t, authChildren), targetReference, true, 100,
	)
	if err != nil {
		t.Fatal(err)
	}
	rootFields := []api.FieldDefinition{
		authUUIDField(t, authRootID, "id", "root_id"),
		authTextField(t, authRootName, "name", "root_name"),
		authDecimalField(t, authRootAmount, "amount", "root_amount"),
	}
	targetFields := []api.FieldDefinition{
		authUUIDField(t, authTargetID, "id", "target_id"),
		authTextField(t, authTargetName, "name", "target_name"),
	}
	root := authCatalog(
		t, rootReference, "auth_root", rootFields,
		[]api.RelationDefinition{children}, authRootID, authRootName,
	)
	target := authCatalog(
		t, targetReference, "auth_target", targetFields,
		[]api.RelationDefinition{}, authTargetID, authTargetName,
	)
	catalogs, err := api.NewCatalogSet(
		rootReference, []api.FieldCatalog{root, target},
	)
	if err != nil {
		t.Fatal(err)
	}
	return authorizationQueryFixture{catalogs: catalogs}
}

func authRelationStorage(t *testing.T, relationID string) api.RelationStorage {
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

func authReference(
	t *testing.T, owner, resource, version, digit string,
) api.CatalogReference {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat(digit, 64))
	reference, err := api.NewCatalogReference(owner, resource, version, hash)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func authCatalog(
	t *testing.T,
	reference api.CatalogReference,
	tableName string,
	fields []api.FieldDefinition,
	relations []api.RelationDefinition,
	stableID string,
	searchID string,
) api.FieldCatalog {
	t.Helper()
	table, _ := sqlidentifier.MintDynamic("frm_" + tableName)
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(true)
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
