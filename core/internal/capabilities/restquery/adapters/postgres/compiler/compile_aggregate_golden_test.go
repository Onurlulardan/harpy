package compiler

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompileAggregateDeterministicGolden(t *testing.T) {
	owner := fixtureAggregateField(t,
		"61111111-1111-4111-8111-111111111111", "owner_id", "fld_owner",
		fixtureUUIDType(t), true, []api.AggregateFunction{api.AggregateFunctionCountDistinct},
		[]api.Operator{api.OperatorEq},
	)
	amount := fixtureAggregateField(t,
		"62222222-2222-4222-8222-222222222222", "amount", "fld_amount",
		fixtureIntegerType(t), false,
		[]api.AggregateFunction{api.AggregateFunctionSum, api.AggregateFunctionAvg},
		[]api.Operator{api.OperatorGTE},
	)
	groupDecl := fixtureAggregateGroup(t, owner, "owner_id")
	measureDecl := fixtureAggregateMeasure(t, "row_count", api.AggregateFunctionCount, nil)
	where := fixturePredicateFilter(t, amount, api.OperatorGTE,
		[]resolved.Value{fixtureScalar(t, api.LogicalTypeInteger, "10")})
	countRef := fixtureAggregateReference(t, api.AggregateReferenceMeasure, 0)
	havingPredicate, _ := resolved.NewHavingPredicate(
		countRef, api.OperatorGTE,
		[]resolved.Value{fixtureScalar(t, api.LogicalTypeInteger, "1")},
	)
	having, _ := resolved.NewPredicateHaving(havingPredicate)
	countSort, _ := resolved.NewAggregateSort(countRef, api.DirectionDescending, false)
	groupRef := fixtureAggregateReference(t, api.AggregateReferenceGroup, 0)
	groupSort, _ := resolved.NewAggregateSort(groupRef, api.DirectionAscending, true)
	query, err := resolved.NewAggregateQuery(
		[]resolved.AggregateGroup{groupDecl}, []resolved.AggregateMeasure{measureDecl},
		&where, []resolved.SearchTerm{}, &having,
		[]resolved.AggregateSort{countSort, groupSort}, 25,
	)
	if err != nil {
		t.Fatal(err)
	}
	visibility := fixturePredicateVisibility(t, owner,
		fixtureScalar(t, api.LogicalTypeUUID, "63333333-3333-4333-8333-333333333333"))
	input := fixtureAggregateInput(t, query,
		[]aggregateGroup{{declaration: groupDecl, field: owner, key: "owner_id"}},
		[]aggregateMeasure{{declaration: measureDecl, key: "row_count", function: api.AggregateFunctionCount}},
		visibility,
	)
	statement, err := compileAggregate(input)
	if err != nil {
		t.Fatal(err)
	}
	want := `WITH "__fc_raw" AS MATERIALIZED (SELECT "fld_owner" AS "__fc_g0", pg_catalog.count(*) AS "__fc_m0" FROM "frm_records" WHERE ("fld_owner" = $1) AND ("fld_amount" >= $2) GROUP BY "fld_owner"), "__fc_typed" AS MATERIALIZED (SELECT "__fc_g0", "__fc_m0" AS "__fc_m0" FROM "__fc_raw"), "__fc_guard" AS MATERIALIZED (SELECT pg_catalog.count(*) AS "__fc_total", pg_catalog.count("__fc_m0") AS "__fc_c0" FROM "__fc_typed") SELECT "__fc_g0", "__fc_m0" FROM "__fc_typed" CROSS JOIN "__fc_guard" WHERE ("__fc_c0" <= "__fc_total") AND ("__fc_m0" >= $3) ORDER BY "__fc_m0" DESC NULLS LAST, "__fc_g0" ASC NULLS LAST LIMIT $4`
	actual, _ := statement.SQL()
	if actual != want {
		t.Fatalf("SQL mismatch\nwant: %s\n got: %s", want, actual)
	}
	arguments, _ := statement.Arguments()
	wantArguments := []any{
		"63333333-3333-4333-8333-333333333333", "10", "1", 25,
	}
	if !reflect.DeepEqual(arguments, wantArguments) {
		t.Fatalf("arguments=%#v", arguments)
	}
}
