package compiler

import (
	"crypto/sha256"
	"reflect"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompileRecordVisibilityAllAndNone(t *testing.T) {
	field, stable, query := fixtureSimpleQuery(t, "display", "display")
	_ = field
	_ = stable
	all := compileFixtureVisibility(t, query, fixtureAllVisibility(t))
	allSQL, _ := all.SQL()
	wantAll := `SELECT "fld_display", "fld_id" FROM "frm_records" ORDER BY "fld_id" ASC NULLS LAST LIMIT $1`
	if allSQL != wantAll {
		t.Fatalf("all visibility SQL=%s", allSQL)
	}
	none, err := resolved.NewNoResourceVisibility(
		sha256.Sum256([]byte("requirement")), fixtureReference(t, "forms"),
		fixturePath(t, ""),
	)
	if err != nil {
		t.Fatal(err)
	}
	denied := compileFixtureVisibility(t, query, none)
	deniedSQL, _ := denied.SQL()
	wantNone := `SELECT "fld_display", "fld_id" FROM "frm_records" WHERE (FALSE) ORDER BY "fld_id" ASC NULLS LAST LIMIT $1`
	if deniedSQL != wantNone {
		t.Fatalf("none visibility SQL=%s", deniedSQL)
	}
}

func TestPresentationNamesNeverReachSQL(t *testing.T) {
	publicAttack := `x" FROM pg_catalog.pg_authid --`
	outputAttack := `json":true,"admin":"yes`
	_, _, query := fixtureSimpleQuery(t, publicAttack, outputAttack)
	statement := compileFixtureVisibility(t, query, fixtureAllVisibility(t))
	sql, err := statement.SQL()
	if err != nil || strings.Contains(sql, publicAttack) || strings.Contains(sql, outputAttack) {
		t.Fatalf("presentation name reached SQL: %s", sql)
	}
	columns, _ := statement.Columns()
	output, present, _ := columns[0].OutputKey()
	if !present || output != outputAttack {
		t.Fatal("output key was not retained as metadata")
	}
}

func TestCompileBooleanClientAndMandatoryFilters(t *testing.T) {
	field := fixtureField(t,
		"6789abcd-6789-4789-8789-6789abcdefab", "state", "fld_state",
		fixtureFoldedTextType(t, false), "fld_state_folded", "",
		[]api.Operator{api.OperatorEq}, false, false,
	)
	a := fixturePredicateFilter(t, field, api.OperatorEq,
		[]resolved.Value{fixtureScalar(t, api.LogicalTypeText, "a")})
	b := fixturePredicateFilter(t, field, api.OperatorEq,
		[]resolved.Value{fixtureScalar(t, api.LogicalTypeText, "b")})
	notB, err := resolved.NewBooleanFilter(api.RequestFilterNot, []resolved.Filter{b})
	if err != nil {
		t.Fatal(err)
	}
	and, err := resolved.NewBooleanFilter(api.RequestFilterAnd, []resolved.Filter{a, notB})
	if err != nil {
		t.Fatal(err)
	}
	binder := newBinder()
	sql, err := compileFilter(and, binder)
	want := `("fld_state" COLLATE pg_catalog.pg_unicode_fast = $1) AND (NOT ("fld_state" COLLATE pg_catalog.pg_unicode_fast = $2))`
	if err != nil || sql != want || !reflect.DeepEqual(binder.arguments, []any{"a", "b"}) {
		t.Fatalf("client filter sql=%s args=%#v err=%v", sql, binder.arguments, err)
	}
	mandatoryA := fixtureMandatoryPredicate(t, field, "a")
	mandatoryB := fixtureMandatoryPredicate(t, field, "b")
	mandatoryOr, err := resolved.NewMandatoryOr(
		[]resolved.MandatoryFilter{mandatoryA, mandatoryB},
	)
	if err != nil {
		t.Fatal(err)
	}
	binder = newBinder()
	sql, err = compileMandatoryFilter(mandatoryOr, binder)
	want = `("fld_state" COLLATE pg_catalog.pg_unicode_fast = $1) OR ("fld_state" COLLATE pg_catalog.pg_unicode_fast = $2)`
	if err != nil || sql != want || !reflect.DeepEqual(binder.arguments, []any{"a", "b"}) {
		t.Fatalf("mandatory sql=%s args=%#v err=%v", sql, binder.arguments, err)
	}
}

func fixtureMandatoryPredicate(
	testingT testing.TB, field api.FieldDefinition, value string,
) resolved.MandatoryFilter {
	testingT.Helper()
	predicate, err := resolved.NewMandatoryPredicate(
		field, api.OperatorEq,
		[]resolved.Value{fixtureScalar(testingT, api.LogicalTypeText, value)},
	)
	if err != nil {
		testingT.Fatal(err)
	}
	filter, err := resolved.NewMandatoryPredicateFilter(predicate)
	if err != nil {
		testingT.Fatal(err)
	}
	return filter
}

func fixtureSimpleQuery(
	testingT testing.TB, publicName, outputKey string,
) (api.FieldDefinition, api.FieldDefinition, resolved.RecordQuery) {
	testingT.Helper()
	field := fixtureField(testingT,
		"789abcde-789a-489a-889a-789abcdefabc", publicName, "fld_display",
		fixtureFoldedTextType(testingT, false), "fld_display_folded", "",
		[]api.Operator{}, false, false,
	)
	stable := fixtureField(testingT,
		"89abcdef-89ab-49ab-89ab-89abcdefabcd", "id", "fld_id",
		fixtureUUIDType(testingT), "", "", []api.Operator{}, true, false,
	)
	query := fixtureRecordQuery(testingT,
		[]resolved.Projection{fixtureProjection(testingT, field, outputKey, false)}, nil,
		[]resolved.Sort{fixtureSort(
			testingT, stable, api.DirectionAscending, true,
		)}, nil, 1,
	)
	return field, stable, query
}

func compileFixtureVisibility(
	testingT testing.TB, query resolved.RecordQuery, visibility resolved.ResourceVisibility,
) RecordStatement {
	testingT.Helper()
	statement, err := compileRecord(
		fixtureRecordInput(testingT, "forms", query, visibility),
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return statement
}
