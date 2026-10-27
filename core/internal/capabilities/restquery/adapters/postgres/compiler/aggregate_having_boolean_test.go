package compiler

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompileAggregateHavingPreservesBooleanASTAndBindsEveryValue(t *testing.T) {
	owner := fixtureAggregateField(t,
		"81111111-1111-4111-8111-111111111111", "owner", "fld_owner",
		fixtureUUIDType(t), true, nil, nil,
	)
	group := fixtureAggregateGroup(t, owner, "owner")
	count := fixtureAggregateMeasure(t, "row_count", api.AggregateFunctionCount, nil)
	groupRef := fixtureAggregateReference(t, api.AggregateReferenceGroup, 0)
	countRef := fixtureAggregateReference(t, api.AggregateReferenceMeasure, 0)
	between := fixtureHavingPredicate(t, countRef, api.OperatorBetween, []resolved.Value{
		fixtureScalar(t, api.LogicalTypeInteger, "1"),
		fixtureScalar(t, api.LogicalTypeInteger, "10"),
	})
	in := fixtureHavingPredicate(t, groupRef, api.OperatorIn, []resolved.Value{
		fixtureScalar(t, api.LogicalTypeUUID, "82222222-2222-4222-8222-222222222222"),
		fixtureScalar(t, api.LogicalTypeUUID, "83333333-3333-4333-8333-333333333333"),
	})
	isNull := fixtureHavingPredicate(t, countRef, api.OperatorIsNull, []resolved.Value{})
	not, _ := resolved.NewBooleanHaving(api.RequestHavingKindNot, []resolved.Having{isNull})
	or, _ := resolved.NewBooleanHaving(api.RequestHavingKindOr, []resolved.Having{in, not})
	having, _ := resolved.NewBooleanHaving(
		api.RequestHavingKindAnd, []resolved.Having{between, or},
	)
	groupSort, _ := resolved.NewAggregateSort(groupRef, api.DirectionAscending, true)
	query, err := resolved.NewAggregateQuery(
		[]resolved.AggregateGroup{group}, []resolved.AggregateMeasure{count}, nil,
		[]resolved.SearchTerm{}, &having, []resolved.AggregateSort{groupSort}, 7,
	)
	if err != nil {
		t.Fatal(err)
	}
	statement, err := compileAggregate(fixtureAggregateInput(
		t, query,
		[]aggregateGroup{{declaration: group, field: owner, key: "owner"}},
		[]aggregateMeasure{{declaration: count, key: "row_count",
			function: api.AggregateFunctionCount}},
		fixtureAllVisibility(t),
	))
	if err != nil {
		t.Fatal(err)
	}
	sql, _ := statement.SQL()
	for _, fragment := range []string{
		`("__fc_m0" BETWEEN $1 AND $2)`,
		`("__fc_g0" IN ($3, $4))`,
		`NOT ("__fc_m0" IS NULL)`,
		`) AND ((`, `) OR (`, `LIMIT $5`,
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("SQL missing %q:\n%s", fragment, sql)
		}
	}
	want := []any{
		"1", "10", "82222222-2222-4222-8222-222222222222",
		"83333333-3333-4333-8333-333333333333", 7,
	}
	arguments, _ := statement.Arguments()
	if !reflect.DeepEqual(arguments, want) {
		t.Fatalf("arguments = %#v", arguments)
	}
}

func fixtureHavingPredicate(
	t testing.TB, reference resolved.AggregateReference,
	operator api.Operator, values []resolved.Value,
) resolved.Having {
	t.Helper()
	predicate, err := resolved.NewHavingPredicate(reference, operator, values)
	if err != nil {
		t.Fatal(err)
	}
	having, err := resolved.NewPredicateHaving(predicate)
	if err != nil {
		t.Fatal(err)
	}
	return having
}
