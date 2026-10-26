package authorization

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestRecordPageDropsProbeRowAndRoundTripsExactCursor(t *testing.T) {
	statement, codec := recordPageStatement(t, true)
	rows := [][]rowdecode.Value{
		recordPageRow(t, statement, "Ada", recordPageUUID(1)),
		recordPageRow(t, statement, "Bob", recordPageUUID(2)),
		recordPageRow(t, statement, "Probe", recordPageUUID(3)),
	}
	page, err := recordpage.Assemble(statement, codec, rows)
	if err != nil {
		t.Fatal(err)
	}
	records, recordsErr := page.Records()
	more, moreErr := page.HasMore()
	token, present, cursorErr := page.NextCursor()
	if recordsErr != nil || moreErr != nil || cursorErr != nil || !more || !present ||
		len(records) != 2 || string(records[0]) != `{"name":"Ada"}` ||
		string(records[1]) != `{"name":"Bob"}` {
		t.Fatalf("records=%q more=%v present=%v errors=%v/%v/%v",
			records, more, present, recordsErr, moreErr, cursorErr)
	}
	if strings.Contains(string(records[1]), "40000000") ||
		strings.Contains(string(records[1]), "Probe") {
		t.Fatal("hidden sort column or probe row leaked")
	}
	context, _ := statement.CursorContext()
	claims, err := codec.Decode(token, context)
	if err != nil {
		t.Fatal(err)
	}
	last, _ := claims.LastValues()
	name, namePresent, _ := last[0].Value()
	nameScalar, scalarPresent, _ := name.Scalar()
	stable, stablePresent, _ := last[1].Value()
	stableScalar, stableScalarPresent, _ := stable.Scalar()
	if !namePresent || !scalarPresent || nameScalar != "Bob" ||
		!stablePresent || !stableScalarPresent ||
		stableScalar != "40000000-0000-4000-8000-000000000002" {
		t.Fatal("cursor did not preserve the final returned sort tuple")
	}
}

func TestRecordPageAllowsNullSortButRejectsNullStableID(t *testing.T) {
	statement, codec := recordPageStatement(t, true)
	rows := [][]rowdecode.Value{
		recordPageRow(t, statement, "Ada", recordPageUUID(1)),
		recordPageRow(t, statement, nil, recordPageUUID(2)),
		recordPageRow(t, statement, "Probe", recordPageUUID(3)),
	}
	page, err := recordpage.Assemble(statement, codec, rows)
	if err != nil {
		t.Fatalf("nullable explicit sort was rejected: %v", err)
	}
	token, _, _ := page.NextCursor()
	context, _ := statement.CursorContext()
	claims, err := codec.Decode(token, context)
	last, _ := claims.LastValues()
	isNull, nullErr := last[0].IsNull()
	if err != nil || nullErr != nil || !isNull {
		t.Fatal("null sort tuple did not round-trip")
	}
	rows[1] = recordPageRow(t, statement, nil, nil)
	if _, err := recordpage.Assemble(statement, codec, rows); err != recordpage.ErrContract {
		t.Fatalf("null stable ID error=%v", err)
	}
}

func TestRecordPageRejectsOverflowWidthAndTypeMismatch(t *testing.T) {
	statement, codec := recordPageStatement(t, false)
	valid := recordPageRow(t, statement, "Ada", recordPageUUID(1))
	probe := recordPageRow(t, statement, "Probe", recordPageUUID(2))
	if _, err := recordpage.Assemble(statement, codec,
		[][]rowdecode.Value{valid, valid, probe, probe}); err != recordpage.ErrContract {
		t.Fatalf("overflow error=%v", err)
	}
	if _, err := recordpage.Assemble(statement, codec,
		[][]rowdecode.Value{valid[:len(valid)-1]}); err != recordpage.ErrContract {
		t.Fatalf("width error=%v", err)
	}
	wrong := append([]rowdecode.Value(nil), valid...)
	uuidType, _ := api.NewUUIDFieldType()
	wrong[0], _ = rowdecode.Decode(uuidType, recordPageUUID(9))
	if _, err := recordpage.Assemble(statement, codec,
		[][]rowdecode.Value{wrong}); err != recordpage.ErrContract {
		t.Fatalf("type error=%v", err)
	}
}

func TestRecordPageEmptyAndShortPagesHaveNoCursor(t *testing.T) {
	statement, codec := recordPageStatement(t, false)
	for name, rows := range map[string][][]rowdecode.Value{
		"empty": nil,
		"short": {recordPageRow(t, statement, "Ada", recordPageUUID(1))},
	} {
		t.Run(name, func(t *testing.T) {
			page, err := recordpage.Assemble(statement, codec, rows)
			if err != nil {
				t.Fatal(err)
			}
			more, _ := page.HasMore()
			cursor, present, _ := page.NextCursor()
			if more || present || cursor != "" {
				t.Fatal("short page exposed continuation")
			}
		})
	}
}
