package executor

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	graphRootResource = "71000000-0000-4000-8000-000000000001"
	graphAResource    = "71000000-0000-4000-8000-000000000002"
	graphBResource    = "71000000-0000-4000-8000-000000000003"
	graphCResource    = "71000000-0000-4000-8000-000000000004"
	graphRootVersion  = "72000000-0000-4000-8000-000000000001"
	graphAVersion     = "72000000-0000-4000-8000-000000000002"
	graphBVersion     = "72000000-0000-4000-8000-000000000003"
	graphCVersion     = "72000000-0000-4000-8000-000000000004"
	graphRootID       = "73000000-0000-4000-8000-000000000001"
	graphRootLabel    = "73000000-0000-4000-8000-000000000002"
	graphAID          = "73000000-0000-4000-8000-000000000003"
	graphALabel       = "73000000-0000-4000-8000-000000000004"
	graphBID          = "73000000-0000-4000-8000-000000000005"
	graphBLabel       = "73000000-0000-4000-8000-000000000006"
	graphCID          = "73000000-0000-4000-8000-000000000007"
	graphCLabel       = "73000000-0000-4000-8000-000000000008"
	graphRelationA    = "74000000-0000-4000-8000-000000000001"
	graphRelationB    = "74000000-0000-4000-8000-000000000002"
	graphRelationC    = "74000000-0000-4000-8000-000000000003"
)

func graphReference(
	t *testing.T,
	owner string,
	resource string,
	version string,
	hashDigit string,
) api.CatalogReference {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat(hashDigit, 64))
	reference, err := api.NewCatalogReference(owner, resource, version, hash)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func graphRelationStorage(t *testing.T, identifier string) api.RelationStorage {
	t.Helper()
	name, err := physicalid.Name("rel_", uuid.MustParse(identifier))
	if err != nil {
		t.Fatal(err)
	}
	token, err := sqlidentifier.MintDynamic(name)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := api.NewRelationStorage(identifier, token)
	if err != nil {
		t.Fatal(err)
	}
	return storage
}
