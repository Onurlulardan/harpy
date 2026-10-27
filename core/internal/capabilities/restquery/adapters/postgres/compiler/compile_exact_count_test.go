package compiler

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompileExactCountDeterministicGolden(t *testing.T) {
	input, _ := fixtureExactCountScenario(t)
	record, recordErr := compileRecord(input)
	count, countErr := compileExactCount(input)
	if recordErr != nil || countErr != nil {
		t.Fatalf("compile record=%v count=%v", recordErr, countErr)
	}
	wantCountSQL := `SELECT COUNT(*) FROM "frm_records" WHERE ("fld_owner" = $1) AND ("fld_name_folded" COLLATE pg_catalog.pg_unicode_fast LIKE $2 ESCAPE E'\\') AND (("fld_name_folded" COLLATE pg_catalog.pg_unicode_fast LIKE $3 ESCAPE E'\\'))`
	countSQL, _ := count.SQL()
	countArguments, _ := count.Arguments()
	wantCountArguments := []any{
		"44444444-5555-4666-8777-888888888888", "%term\\%\\_%", "%needle%",
	}
	if countSQL != wantCountSQL ||
		!reflect.DeepEqual(countArguments, wantCountArguments) {
		t.Fatalf("count SQL=%s args=%#v", countSQL, countArguments)
	}
	recordSQL, _ := record.SQL()
	recordArguments, _ := record.Arguments()
	if !strings.Contains(recordSQL, ` ORDER BY "fld_id" ASC NULLS LAST LIMIT $4`) ||
		!reflect.DeepEqual(
			recordArguments, append(wantCountArguments, 3),
		) || strings.Contains(countSQL, "ORDER BY") || strings.Contains(countSQL, "LIMIT") {
		t.Fatalf("record/count semantic drift: record=%s count=%s", recordSQL, countSQL)
	}
	explain, err := count.ExplainSQL()
	if err != nil || explain != explainJSONPrefix+wantCountSQL {
		t.Fatalf("exact EXPLAIN SQL=%s err=%v", explain, err)
	}
}

func TestCompileExactCountIgnoresValidatedContinuationTuple(t *testing.T) {
	input, _ := fixtureExactCountScenario(t)
	value := fixtureScalar(
		t, api.LogicalTypeUUID, "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee",
	)
	last, err := cursor.NewLastValue(value)
	if err != nil {
		t.Fatal(err)
	}
	input.last = []cursor.LastValue{last}
	statement, err := compileExactCount(input)
	sql, _ := statement.SQL()
	arguments, _ := statement.Arguments()
	if err != nil || strings.Contains(sql, `"fld_id" >`) || len(arguments) != 3 {
		t.Fatalf("continuation leaked into exact count: %s %#v %v", sql, arguments, err)
	}
}

func TestCompileExactCountRejectsCountNone(t *testing.T) {
	_, _, query := fixtureSimpleQuery(t, "display", "display")
	_, err := compileExactCount(fixtureRecordInput(
		t, "forms", query, fixtureAllVisibility(t),
	))
	if err != ErrContract {
		t.Fatalf("count=none exact compile error=%v", err)
	}
}

func fixtureExactQueryWithValue(
	testingT testing.TB, raw string,
) recordInput {
	testingT.Helper()
	field := fixtureField(testingT,
		"55555555-6666-4777-8888-999999999999", "text", "fld_text",
		fixtureFoldedTextType(testingT, false), "fld_text_folded", "",
		[]api.Operator{api.OperatorEq}, false, false,
	)
	stable := fixtureField(testingT,
		"66666666-7777-4888-8999-aaaaaaaaaaaa", "id", "fld_id",
		fixtureUUIDType(testingT), "", "", []api.Operator{}, true, false,
	)
	where := fixturePredicateFilter(testingT, field, api.OperatorEq,
		[]resolved.Value{fixtureScalar(testingT, api.LogicalTypeText, raw)})
	page, _ := resolved.NewPage(1, nil)
	query, err := resolved.NewRecordQuery(
		[]resolved.Projection{fixtureProjection(testingT, field, "text", false)},
		&where, []resolved.Sort{fixtureSort(
			testingT, stable, api.DirectionAscending, true,
		)}, []resolved.SearchTerm{}, page, api.CountModeExact,
		fixturePath(testingT, "/count"), true, []resolved.Include{},
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return fixtureRecordInput(testingT, "forms", query, fixtureAllVisibility(testingT))
}
