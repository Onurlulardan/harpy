package authorization

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

const (
	testFieldID    = "30000000-0000-4000-8000-000000000001"
	testRelationID = "40000000-0000-4000-8000-000000000001"
)

func requirementCatalog(t *testing.T) api.CatalogReference {
	t.Helper()
	hash, err := api.NewCatalogHash(strings.Repeat("a", 64))
	if err != nil {
		t.Fatal(err)
	}
	reference, err := api.NewCatalogReference(
		"forms",
		"10000000-0000-4000-8000-000000000001",
		"20000000-0000-4000-8000-000000000001",
		hash,
	)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func requirementPath(t *testing.T, value string) resolved.RequestPath {
	t.Helper()
	path, err := resolved.NewRequestPath(value)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func mustRequirement(
	t *testing.T,
	kind RequirementKind,
	resourcePath string,
	requestPath string,
	fieldID uuid.UUID,
	relationID uuid.UUID,
) Requirement {
	t.Helper()
	requirement, err := newRequirement(
		kind,
		requirementCatalog(t),
		requirementPath(t, resourcePath),
		requirementPath(t, requestPath),
		fieldID,
		relationID,
	)
	if err != nil {
		t.Fatal(err)
	}
	return requirement
}
