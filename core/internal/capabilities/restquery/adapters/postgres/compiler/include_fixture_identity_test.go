package compiler

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	includeRootResource = "60000000-0000-4000-8000-000000000001"
	includeAResource    = "60000000-0000-4000-8000-000000000002"
	includeBResource    = "60000000-0000-4000-8000-000000000003"
	includeCResource    = "60000000-0000-4000-8000-000000000004"
	includeRootVersion  = "61000000-0000-4000-8000-000000000001"
	includeAVersion     = "61000000-0000-4000-8000-000000000002"
	includeBVersion     = "61000000-0000-4000-8000-000000000003"
	includeCVersion     = "61000000-0000-4000-8000-000000000004"
	includeRootID       = "62000000-0000-4000-8000-000000000001"
	includeRootLabel    = "62000000-0000-4000-8000-000000000002"
	includeAID          = "62000000-0000-4000-8000-000000000003"
	includeALabel       = "62000000-0000-4000-8000-000000000004"
	includeBID          = "62000000-0000-4000-8000-000000000005"
	includeBLabel       = "62000000-0000-4000-8000-000000000006"
	includeCID          = "62000000-0000-4000-8000-000000000007"
	includeCLabel       = "62000000-0000-4000-8000-000000000008"
	includeRelationA    = "63000000-0000-4000-8000-000000000001"
	includeRelationB    = "63000000-0000-4000-8000-000000000002"
	includeRelationC    = "63000000-0000-4000-8000-000000000003"
)

func includeReference(
	t *testing.T, owner, resource, version, hashDigit string,
) api.CatalogReference {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat(hashDigit, 64))
	reference, err := api.NewCatalogReference(owner, resource, version, hash)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func includeRelationStorage(t *testing.T, identifier string) api.RelationStorage {
	t.Helper()
	name, err := physicalid.Name("rel_", uuid.MustParse(identifier))
	if err != nil {
		t.Fatal(err)
	}
	table, err := sqlidentifier.MintDynamic(name)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := api.NewRelationStorage(identifier, table)
	if err != nil {
		t.Fatal(err)
	}
	return storage
}
