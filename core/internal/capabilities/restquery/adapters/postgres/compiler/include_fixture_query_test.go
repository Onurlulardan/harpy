package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

const includeInjectionValue = `x' OR true; DROP TABLE frm_include_a; --`

func includeProjection(
	t *testing.T, field api.FieldDefinition, key, path string,
) resolved.Projection {
	t.Helper()
	output, err := resolved.NewOutputKey(key)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := resolved.NewProjection(
		field, output, fixturePath(t, path), false,
	)
	if err != nil {
		t.Fatal(err)
	}
	return projection
}

func includeSort(
	t *testing.T, field api.FieldDefinition, direction api.Direction,
	path string, implicit bool,
) resolved.Sort {
	t.Helper()
	sort, err := resolved.NewSort(field, direction, fixturePath(t, path), implicit)
	if err != nil {
		t.Fatal(err)
	}
	return sort
}

func includePredicate(
	t *testing.T, field api.FieldDefinition, value, path string,
) resolved.Filter {
	t.Helper()
	resolvedValue, err := resolved.NewScalarValue(api.LogicalTypeText, value)
	if err != nil {
		t.Fatal(err)
	}
	predicate, err := resolved.NewPredicate(
		field, api.OperatorEq, []resolved.Value{resolvedValue}, fixturePath(t, path),
	)
	if err != nil {
		t.Fatal(err)
	}
	filter, err := resolved.NewPredicateFilter(predicate)
	if err != nil {
		t.Fatal(err)
	}
	return filter
}

func includeResolvedNode(
	t *testing.T,
	relation api.RelationDefinition,
	target api.CatalogReference,
	storage api.ResourceStorage,
	stable api.FieldDefinition,
	sortField api.FieldDefinition,
	base string,
	where *resolved.Filter,
	children []resolved.Include,
) resolved.Include {
	t.Helper()
	relationStorage, _ := relation.Storage()
	projections := []resolved.Projection{
		includeProjection(t, stable, "id", base+"/select"),
	}
	sorts := []resolved.Sort{
		includeSort(t, sortField, api.DirectionAscending, base+"/sort/0", false),
		includeSort(t, stable, api.DirectionAscending, base+"/sort", true),
	}
	include, err := resolved.NewInclude(
		relation, target, storage, relationStorage, stable, fixturePath(t, base),
		projections, where, sorts, 2, children,
	)
	if err != nil {
		t.Fatal(err)
	}
	return include
}
