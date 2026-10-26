package compiler

import (
	"crypto/sha256"
	"math"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func fixtureIntegerType(t testing.TB) api.FieldType {
	t.Helper()
	profile, err := api.NewIntegerProfile(math.MinInt64, math.MaxInt64)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewIntegerFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func fixtureExactTextType(t testing.TB) api.FieldType {
	t.Helper()
	profile, err := api.NewTextProfile(
		api.TextCasefoldDisabled, api.TextAccentDisabled,
	)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewTextFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func fixtureAggregateField(
	t testing.TB, id, name, column string, fieldType api.FieldType,
	groupable bool, aggregates []api.AggregateFunction,
	operators []api.Operator,
) api.FieldDefinition {
	t.Helper()
	storage, err := api.NewFieldStorage(fixtureToken(t, column), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	filters := make([]api.FilterCapabilityInput, len(operators))
	for index, operator := range operators {
		filters[index] = api.FilterCapabilityInput{Operator: operator, CountSafe: true}
	}
	capabilities, err := api.NewFieldCapabilities(fieldType, api.FieldCapabilitiesInput{
		Selectable: true, Filters: filters, Groupable: groupable,
		Aggregates: aggregates,
	})
	if err != nil {
		t.Fatal(err)
	}
	field, err := api.NewFieldDefinition(id, name, fieldType, storage, capabilities)
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func fixtureAggregateGroup(
	t testing.TB, field api.FieldDefinition, key string,
) resolved.AggregateGroup {
	t.Helper()
	output, _ := resolved.NewOutputKey(key)
	group, err := resolved.NewAggregateGroup(field, output, fixturePath(t, "/aggregate/group_by/0"))
	if err != nil {
		t.Fatal(err)
	}
	return group
}

func fixtureAggregateMeasure(
	t testing.TB, key string, function api.AggregateFunction,
	field *api.FieldDefinition,
) resolved.AggregateMeasure {
	t.Helper()
	output, _ := resolved.NewOutputKey(key)
	measure, err := resolved.NewAggregateMeasure(
		output, function, field, fixturePath(t, "/aggregate/measures/0"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return measure
}

func fixtureAggregateReference(
	t testing.TB, kind api.AggregateReference, ordinal int,
) resolved.AggregateReference {
	t.Helper()
	reference, err := resolved.NewAggregateReference(
		kind, ordinal, fixturePath(t, "/aggregate/reference"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func fixtureAggregateInput(
	t testing.TB, query resolved.AggregateQuery,
	groups []aggregateGroup, measures []aggregateMeasure,
	visibility resolved.ResourceVisibility,
) aggregateInput {
	t.Helper()
	binding, err := newStatementBinding(
		fixtureCursorBinding(t, "forms"), aggregateContinuation,
	)
	if err != nil {
		t.Fatal(err)
	}
	return aggregateInput{
		table: fixtureToken(t, "frm_records"), query: query,
		groups: groups, measures: measures,
		visibility: visibility, binding: binding,
	}
}

func fixturePredicateVisibility(
	t testing.TB, field api.FieldDefinition, value resolved.Value,
) resolved.ResourceVisibility {
	t.Helper()
	predicate, _ := resolved.NewMandatoryPredicate(field, api.OperatorEq, []resolved.Value{value})
	filter, _ := resolved.NewMandatoryPredicateFilter(predicate)
	visibility, err := resolved.NewPredicateResourceVisibility(
		sha256.Sum256([]byte("aggregate-requirement")), fixtureReference(t, "forms"),
		fixturePath(t, ""), filter,
	)
	if err != nil {
		t.Fatal(err)
	}
	return visibility
}
