package compiler

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompilePredicateOperatorMatrix(t *testing.T) {
	operators := []api.Operator{
		api.OperatorEq, api.OperatorNe, api.OperatorGT, api.OperatorGTE,
		api.OperatorLT, api.OperatorLTE, api.OperatorContains,
		api.OperatorStartsWith, api.OperatorEndsWith, api.OperatorIn,
		api.OperatorNotIn, api.OperatorBetween, api.OperatorIsNull,
		api.OperatorIsNotNull,
	}
	field := fixtureField(t,
		"12345678-1234-4234-8234-123456789abc", "text", "fld_text",
		fixtureFoldedTextType(t, false), "fld_text_folded", "", operators, true, true,
	)
	a := fixtureScalar(t, api.LogicalTypeText, "a")
	b := fixtureScalar(t, api.LogicalTypeText, "b")
	pattern := fixtureScalar(t, api.LogicalTypeText, `a%_\b`)
	exact := `"fld_text" COLLATE pg_catalog.pg_unicode_fast`
	patternColumn := `"fld_text_folded" COLLATE pg_catalog.pg_unicode_fast`
	tests := []struct {
		name     string
		operator api.Operator
		values   []resolved.Value
		sql      string
		args     []any
	}{
		{"eq", api.OperatorEq, []resolved.Value{a}, exact + ` = $1`, []any{"a"}},
		{"ne", api.OperatorNe, []resolved.Value{a}, exact + ` <> $1`, []any{"a"}},
		{"gt", api.OperatorGT, []resolved.Value{a}, exact + ` > $1`, []any{"a"}},
		{"gte", api.OperatorGTE, []resolved.Value{a}, exact + ` >= $1`, []any{"a"}},
		{"lt", api.OperatorLT, []resolved.Value{a}, exact + ` < $1`, []any{"a"}},
		{"lte", api.OperatorLTE, []resolved.Value{a}, exact + ` <= $1`, []any{"a"}},
		{"contains", api.OperatorContains, []resolved.Value{pattern}, patternColumn +
			` LIKE $1 ESCAPE E'\\'`, []any{`%a\%\_\\b%`}},
		{"starts", api.OperatorStartsWith, []resolved.Value{pattern}, patternColumn +
			` LIKE $1 ESCAPE E'\\'`, []any{`a\%\_\\b%`}},
		{"ends", api.OperatorEndsWith, []resolved.Value{pattern}, patternColumn +
			` LIKE $1 ESCAPE E'\\'`, []any{`%a\%\_\\b`}},
		{"in", api.OperatorIn, []resolved.Value{a, b}, exact +
			` IN ($1, $2)`, []any{"a", "b"}},
		{"not-in", api.OperatorNotIn, []resolved.Value{a, b}, exact +
			` NOT IN ($1, $2)`, []any{"a", "b"}},
		{"between", api.OperatorBetween, []resolved.Value{a, b}, exact +
			` BETWEEN $1 AND $2`, []any{"a", "b"}},
		{"null", api.OperatorIsNull, []resolved.Value{}, `"fld_text" IS NULL`, []any{}},
		{"not-null", api.OperatorIsNotNull, []resolved.Value{},
			`"fld_text" IS NOT NULL`, []any{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			binder := newBinder()
			actual, err := compilePredicate(field, test.operator, test.values, binder)
			if err != nil || actual != test.sql ||
				!reflect.DeepEqual(binder.arguments, test.args) {
				t.Fatalf("sql=%s args=%#v err=%v", actual, binder.arguments, err)
			}
		})
	}
}

func TestCompileCollectionAndJSONPredicates(t *testing.T) {
	textList := fixtureField(t,
		"23456789-2345-4345-8345-23456789abcd", "tags", "fld_tags",
		fixtureTextListType(t), "", "",
		[]api.Operator{api.OperatorContainsAny, api.OperatorContainsAll}, false, false,
	)
	values := []resolved.Value{
		fixtureScalar(t, api.LogicalTypeText, "blue"),
		fixtureScalar(t, api.LogicalTypeText, "green"),
	}
	for _, test := range []struct {
		operator  api.Operator
		operation string
	}{{api.OperatorContainsAny, " && "}, {api.OperatorContainsAll, " @> "}} {
		binder := newBinder()
		sql, err := compilePredicate(textList, test.operator, values, binder)
		want := `"fld_tags"` + test.operation +
			`CAST(ARRAY[$1, $2] AS pg_catalog.text[])`
		if err != nil || sql != want ||
			!reflect.DeepEqual(binder.arguments, []any{"blue", "green"}) {
			t.Fatalf("collection sql=%s args=%#v err=%v", sql, binder.arguments, err)
		}
	}
	jsonField := fixtureField(t,
		"3456789a-3456-4456-8456-3456789abcde", "payload", "fld_payload",
		fixtureJSONType(t), "", "", []api.Operator{api.OperatorEq}, false, false,
	)
	document, err := resolved.NewJSONValue([]byte(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	binder := newBinder()
	sql, err := compilePredicate(
		jsonField, api.OperatorEq, []resolved.Value{document}, binder,
	)
	if err != nil || sql != `"fld_payload" = CAST($1 AS pg_catalog.jsonb)` ||
		!reflect.DeepEqual(binder.arguments, []any{`{"a":1}`}) {
		t.Fatalf("JSON sql=%s args=%#v err=%v", sql, binder.arguments, err)
	}
}
