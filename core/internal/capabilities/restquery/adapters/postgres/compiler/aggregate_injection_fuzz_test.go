package compiler

import (
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"golang.org/x/text/unicode/norm"
)

func FuzzCompileAggregateNeverInterpolatesOutputOrValue(f *testing.F) {
	f.Add(`x"; DROP TABLE users; --`, `x' OR TRUE --`)
	f.Add("display", "ordinary")
	field := fixtureAggregateField(f,
		"81111111-1111-4111-8111-111111111111", `public";drop`, "fld_group",
		fixtureExactTextType(f), true, nil, nil,
	)
	const expectedSQL = `WITH "__fc_raw" AS MATERIALIZED (SELECT "fld_group" COLLATE pg_catalog.pg_unicode_fast AS "__fc_g0" FROM "frm_records" GROUP BY "fld_group" COLLATE pg_catalog.pg_unicode_fast), "__fc_typed" AS MATERIALIZED (SELECT "__fc_g0" FROM "__fc_raw"), "__fc_guard" AS MATERIALIZED (SELECT pg_catalog.count(*) AS "__fc_total" FROM "__fc_typed") SELECT "__fc_g0" FROM "__fc_typed" CROSS JOIN "__fc_guard" WHERE ("__fc_total" >= 0) AND ("__fc_g0" COLLATE pg_catalog.pg_unicode_fast = $1) ORDER BY "__fc_g0" COLLATE pg_catalog.pg_unicode_fast ASC NULLS LAST LIMIT $2`
	f.Fuzz(func(t *testing.T, outputKey, value string) {
		if outputKey == "" || outputKey == "_included" || !utf8.ValidString(outputKey) ||
			!norm.NFC.IsNormalString(outputKey) || !utf8.ValidString(value) ||
			!norm.NFC.IsNormalString(value) {
			t.Skip()
		}
		output, err := resolved.NewOutputKey(outputKey)
		if err != nil {
			t.Skip()
		}
		group, err := resolved.NewAggregateGroup(
			field, output, fixturePath(t, "/aggregate/group_by/0"),
		)
		if err != nil {
			t.Fatal(err)
		}
		resolvedValue, err := resolved.NewScalarValue(api.LogicalTypeText, value)
		if err != nil {
			t.Skip()
		}
		reference := fixtureAggregateReference(t, api.AggregateReferenceGroup, 0)
		predicate, _ := resolved.NewHavingPredicate(
			reference, api.OperatorEq, []resolved.Value{resolvedValue},
		)
		having, _ := resolved.NewPredicateHaving(predicate)
		sort, _ := resolved.NewAggregateSort(reference, api.DirectionAscending, true)
		query, err := resolved.NewAggregateQuery(
			[]resolved.AggregateGroup{group}, []resolved.AggregateMeasure{}, nil,
			[]resolved.SearchTerm{}, &having, []resolved.AggregateSort{sort}, 100,
		)
		if err != nil {
			t.Fatal(err)
		}
		groups, _ := extractAggregateGroups(query)
		statement, err := compileAggregate(fixtureAggregateInput(
			t, query, groups, []aggregateMeasure{}, fixtureAllVisibility(t),
		))
		if err != nil {
			t.Fatal(err)
		}
		sql, _ := statement.SQL()
		arguments, _ := statement.Arguments()
		if sql != expectedSQL || !reflect.DeepEqual(arguments, []any{value, 100}) {
			t.Fatalf("SQL=%s args=%#v", sql, arguments)
		}
	})
}
