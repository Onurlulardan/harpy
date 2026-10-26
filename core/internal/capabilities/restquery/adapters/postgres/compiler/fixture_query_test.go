package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func fixtureProjection(
	testingT testing.TB, field api.FieldDefinition, key string, hidden bool,
) resolved.Projection {
	testingT.Helper()
	output, err := resolved.NewOutputKey(key)
	if err != nil {
		testingT.Fatal(err)
	}
	projection, err := resolved.NewProjection(
		field, output, fixturePath(testingT, "/select"), hidden,
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return projection
}

func fixtureSort(
	testingT testing.TB,
	field api.FieldDefinition,
	direction api.Direction,
	implicit bool,
) resolved.Sort {
	testingT.Helper()
	sortValue, err := resolved.NewSort(
		field, direction, fixturePath(testingT, "/sort"), implicit,
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return sortValue
}

func fixtureScalar(
	testingT testing.TB, kind api.LogicalType, value string,
) resolved.Value {
	testingT.Helper()
	result, err := resolved.NewScalarValue(kind, value)
	if err != nil {
		testingT.Fatal(err)
	}
	return result
}

func fixturePredicateFilter(
	testingT testing.TB,
	field api.FieldDefinition,
	operator api.Operator,
	values []resolved.Value,
) resolved.Filter {
	testingT.Helper()
	predicate, err := resolved.NewPredicate(
		field, operator, values, fixturePath(testingT, "/filter"),
	)
	if err != nil {
		testingT.Fatal(err)
	}
	filter, err := resolved.NewPredicateFilter(predicate)
	if err != nil {
		testingT.Fatal(err)
	}
	return filter
}

func fixtureRecordQuery(
	testingT testing.TB,
	projections []resolved.Projection,
	where *resolved.Filter,
	sorts []resolved.Sort,
	search []resolved.SearchTerm,
	pageSize int,
) resolved.RecordQuery {
	testingT.Helper()
	page, err := resolved.NewPage(pageSize, nil)
	if err != nil {
		testingT.Fatal(err)
	}
	query, err := resolved.NewRecordQuery(
		projections, where, sorts, search, page, api.CountModeNone,
		fixturePath(testingT, "/count"), false, []resolved.Include{},
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return query
}
