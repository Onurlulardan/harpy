package compiler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestV1PredecessorAndDirectKeysetGolden(t *testing.T) {
	_, _, query := fixtureSimpleQuery(t, "display", "display")
	input := fixtureRecordInput(t, "forms", query, fixtureAllVisibility(t))
	columns, err := v1PredecessorColumns(query)
	if err != nil {
		t.Fatal(err)
	}
	binder := newBinder()
	sql, err := compileV1PredecessorSQL(input, columns, binder, 3)
	wantSQL := `SELECT "fld_id" FROM "frm_records" ORDER BY "fld_id" ASC NULLS LAST LIMIT $1 OFFSET $2`
	if err != nil || sql != wantSQL ||
		!reflect.DeepEqual(binder.arguments, []any{1, int64(3)}) {
		t.Fatalf("SQL=%s arguments=%#v error=%v", sql, binder.arguments, err)
	}
	binding, _ := v1PredecessorBinding(input.binding, 3)
	predecessor, err := newPredecessorStatement(
		sql, binder.arguments, resultColumns(columns), 3,
		planner.StatementClassV1PredecessorSeek, binding,
	)
	if err != nil {
		t.Fatal(err)
	}
	last := fixtureUUIDLast(t, "40000000-0000-4000-8000-000000000004")
	tuple, err := predecessor.MintTuple([]cursor.LastValue{last})
	if err != nil {
		t.Fatal(err)
	}
	input.last = []cursor.LastValue{last}
	input.binding, _ = v1PageBinding(input.binding, predecessor.stateSeal, tuple.stateSeal)
	page, err := compileRecord(input)
	pageSQL, _ := page.SQL()
	if err != nil || page.Validate() != nil ||
		pageSQL != `SELECT "fld_display", "fld_id" FROM "frm_records" WHERE ((("fld_id" > $1 OR "fld_id" IS NULL))) ORDER BY "fld_id" ASC NULLS LAST LIMIT $2` {
		t.Fatalf("page SQL=%s error=%v", pageSQL, err)
	}
}

func fixtureUUIDLast(t *testing.T, value string) cursor.LastValue {
	t.Helper()
	resolvedValue, err := resolved.NewScalarValue(api.LogicalTypeUUID, value)
	if err != nil {
		t.Fatal(err)
	}
	last, err := cursor.NewLastValue(resolvedValue)
	if err != nil {
		t.Fatal(err)
	}
	return last
}

func TestV1PredecessorTupleRejectsTransplantTamperAndSerialization(t *testing.T) {
	_, _, query := fixtureSimpleQuery(t, "display", "display")
	input := fixtureRecordInput(t, "forms", query, fixtureAllVisibility(t))
	columns, _ := v1PredecessorColumns(query)
	first := fixturePredecessorStatement(t, input, columns, 1)
	second := fixturePredecessorStatement(t, input, columns, 2)
	tuple, _ := first.MintTuple([]cursor.LastValue{
		fixtureUUIDLast(t, "40000000-0000-4000-8000-000000000001"),
	})
	if tuple.validateFor(second) == nil {
		t.Fatal("ordinal tuple transplant was accepted")
	}
	tampered := tuple
	tampered.stateSeal[0] ^= 0xff
	if tampered.Validate() == nil {
		t.Fatal("tuple seal tamper was accepted")
	}
	if fmt.Sprint(first) != compilerRedaction || fmt.Sprint(tuple) != compilerRedaction {
		t.Fatal("V1 compiler capability formatting leaked")
	}
	for _, value := range []any{first, tuple} {
		if document, err := json.Marshal(value); err == nil || document != nil {
			t.Fatal("V1 compiler capability JSON marshal did not fail closed")
		}
	}
}
