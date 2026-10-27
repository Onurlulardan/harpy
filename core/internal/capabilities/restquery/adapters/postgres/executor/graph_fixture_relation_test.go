package executor

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func graphRelation(
	t *testing.T,
	identifier string,
	name string,
	cardinality api.RelationCardinality,
	target api.CatalogReference,
) api.RelationDefinition {
	t.Helper()
	maximum := 100
	if cardinality == api.RelationCardinalityToOne {
		maximum = 1
	}
	relation, err := api.NewRelationDefinition(
		identifier, name, cardinality, graphRelationStorage(t, identifier),
		target, true, maximum,
	)
	if err != nil {
		t.Fatal(err)
	}
	return relation
}
