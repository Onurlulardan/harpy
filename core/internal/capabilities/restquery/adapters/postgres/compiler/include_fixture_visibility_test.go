package compiler

import (
	"crypto/sha256"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func includeVisibilities(
	t *testing.T, catalogs includeCatalogFixture,
) []resolved.ResourceVisibility {
	t.Helper()
	return []resolved.ResourceVisibility{
		includeAllVisibility(t, catalogs.root, "", "root"),
		includePredicateVisibility(t, catalogs, "/include/0"),
		includeAllVisibility(t, catalogs.targetB, "/include/1", "target-b"),
		includeAllVisibility(
			t, catalogs.targetC, "/include/0/include/0", "target-c",
		),
	}
}

func includeAllVisibility(
	t *testing.T, catalog api.CatalogReference, path, requirement string,
) resolved.ResourceVisibility {
	t.Helper()
	visibility, err := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("include-"+requirement)),
		catalog, fixturePath(t, path),
	)
	if err != nil {
		t.Fatal(err)
	}
	return visibility
}

func includePredicateVisibility(
	t *testing.T, catalogs includeCatalogFixture, path string,
) resolved.ResourceVisibility {
	t.Helper()
	value, err := resolved.NewScalarValue(api.LogicalTypeText, "blocked")
	if err != nil {
		t.Fatal(err)
	}
	predicate, err := resolved.NewMandatoryPredicate(
		catalogs.aLabel, api.OperatorNe, []resolved.Value{value},
	)
	if err != nil {
		t.Fatal(err)
	}
	filter, err := resolved.NewMandatoryPredicateFilter(predicate)
	if err != nil {
		t.Fatal(err)
	}
	visibility, err := resolved.NewPredicateResourceVisibility(
		sha256.Sum256([]byte("include-target-a")), catalogs.targetA,
		fixturePath(t, path), filter,
	)
	if err != nil {
		t.Fatal(err)
	}
	return visibility
}
