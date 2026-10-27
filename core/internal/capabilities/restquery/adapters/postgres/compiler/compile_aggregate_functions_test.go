package compiler

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompileAggregateFunctionsUseExactTypedPipeline(t *testing.T) {
	owner := fixtureAggregateField(t,
		"71111111-1111-4111-8111-111111111111", "owner", "fld_owner",
		fixtureUUIDType(t), false,
		[]api.AggregateFunction{api.AggregateFunctionCountDistinct}, nil,
	)
	integer := fixtureAggregateField(t,
		"72222222-2222-4222-8222-222222222222", "quantity", "fld_quantity",
		fixtureIntegerType(t), false,
		[]api.AggregateFunction{api.AggregateFunctionSum, api.AggregateFunctionAvg}, nil,
	)
	decimal := fixtureAggregateField(t,
		"73333333-3333-4333-8333-333333333333", "amount", "fld_amount",
		fixtureDecimalType(t, 12, 2), false,
		[]api.AggregateFunction{api.AggregateFunctionSum, api.AggregateFunctionAvg}, nil,
	)
	enum := fixtureAggregateField(t,
		"74444444-4444-4444-8444-444444444444", "status", "fld_status",
		fixtureOrderedEnumType(t), false,
		[]api.AggregateFunction{api.AggregateFunctionMin}, nil,
	)
	text := fixtureAggregateField(t,
		"75555555-5555-4555-8555-555555555555", "title", "fld_title",
		fixtureExactTextType(t), false,
		[]api.AggregateFunction{api.AggregateFunctionMax}, nil,
	)
	declarations := []resolved.AggregateMeasure{
		fixtureAggregateMeasure(t, "rows", api.AggregateFunctionCount, nil),
		fixtureAggregateMeasure(t, "owners", api.AggregateFunctionCountDistinct, &owner),
		fixtureAggregateMeasure(t, "quantity_sum", api.AggregateFunctionSum, &integer),
		fixtureAggregateMeasure(t, "quantity_avg", api.AggregateFunctionAvg, &integer),
		fixtureAggregateMeasure(t, "amount_sum", api.AggregateFunctionSum, &decimal),
		fixtureAggregateMeasure(t, "amount_avg", api.AggregateFunctionAvg, &decimal),
		fixtureAggregateMeasure(t, "first_status", api.AggregateFunctionMin, &enum),
		fixtureAggregateMeasure(t, "last_title", api.AggregateFunctionMax, &text),
	}
	query, err := resolved.NewAggregateQuery(
		[]resolved.AggregateGroup{}, declarations, nil,
		[]resolved.SearchTerm{}, nil, []resolved.AggregateSort{}, 100,
	)
	if err != nil {
		t.Fatal(err)
	}
	measures, err := extractAggregateMeasures(query)
	if err != nil {
		t.Fatal(err)
	}
	statement, err := compileAggregate(fixtureAggregateInput(
		t, query, []aggregateGroup{}, measures, fixtureAllVisibility(t),
	))
	if err != nil {
		t.Fatal(err)
	}
	sql, _ := statement.SQL()
	for _, fragment := range []string{
		`pg_catalog.count(*)`,
		`pg_catalog.count(DISTINCT "fld_owner")`,
		`pg_catalog.sum(CAST("fld_quantity" AS pg_catalog.numeric))`,
		`pg_catalog.count("fld_quantity") AS "__fc_n3"`,
		`CAST("__fc_m2" AS pg_catalog.int8)`,
		`AS pg_catalog.numeric(29,6)`,
		`AS pg_catalog.numeric(22,2)`,
		`AS pg_catalog.numeric(22,6)`,
		`pg_catalog.div("__fc_s3", "__fc_d3") AS "__fc_q3"`,
		`pg_catalog.mod(pg_catalog.abs("__fc_q3"), 2)`,
		`(CAST($1 AS pg_catalog.text[]))[pg_catalog.min(`,
		`pg_catalog.max("fld_title" COLLATE pg_catalog.pg_unicode_fast)`,
		`pg_catalog.count("__fc_m7") AS "__fc_c7"`,
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("SQL missing %q:\n%s", fragment, sql)
		}
	}
	if strings.Contains(sql, "pg_catalog.round(") || strings.Contains(sql, "float") ||
		strings.Contains(sql, "pg_catalog.avg(") {
		t.Fatalf("SQL used forbidden numeric path: %s", sql)
	}
	arguments, _ := statement.Arguments()
	wantArguments := []any{[]string{"OPEN", "CLOSED"}, 6, 2, 6, 100}
	if !reflect.DeepEqual(arguments, wantArguments) {
		t.Fatalf("arguments=%#v", arguments)
	}
}
