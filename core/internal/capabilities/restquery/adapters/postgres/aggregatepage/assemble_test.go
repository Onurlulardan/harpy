package aggregatepage

import (
	"math/big"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAssemblePreservesDeclarationOrderNullsAndOwnedCopies(t *testing.T) {
	statement := fixtureAggregateStatement(t)
	rows := [][]rowdecode.Value{
		fixtureAggregateRow(t, statement, "OPEN", int64(3), pgtype.Numeric{
			Int: big.NewInt(1234), Exp: -2, Valid: true,
		}),
		fixtureAggregateRow(t, statement, nil, int64(0), nil),
	}
	page, err := Assemble(statement, rows)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := page.Groups()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		`{"key":{"status":"OPEN"},"measures":{"row_count":"3","revenue":"12.34"}}`,
		`{"key":{"status":null},"measures":{"row_count":"0","revenue":null}}`,
	}
	for index := range expected {
		if string(groups[index]) != expected[index] {
			t.Fatalf("group %d = %s", index, groups[index])
		}
	}
	groups[0][0] = '['
	again, _ := page.Groups()
	if string(again[0]) != expected[0] {
		t.Fatal("caller mutation changed sealed page")
	}
	if limit, err := page.Limit(); err != nil || limit != 2 {
		t.Fatalf("limit = %d, %v", limit, err)
	}
}

func TestAssembleRejectsRowShapeTypeAndInvalidInputs(t *testing.T) {
	statement := fixtureAggregateStatement(t)
	valid := fixtureAggregateRow(t, statement, "OPEN", int64(1), nil)
	metadata, _ := loadMetadata(statement)
	wrongType, _ := rowdecode.Decode(metadata.groups[0].fieldType, "not-an-integer")
	for name, rows := range map[string][][]rowdecode.Value{
		"nil rows":   nil,
		"short row":  {valid[:len(valid)-1]},
		"wrong type": {{valid[0], wrongType, valid[2]}},
		"zero value": {{valid[0], {}, valid[2]}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Assemble(statement, rows); err != ErrContract {
				t.Fatalf("error = %v", err)
			}
		})
	}
	if _, err := Assemble(compiler.AggregateStatement{}, [][]rowdecode.Value{}); err != ErrContract {
		t.Fatalf("zero statement error = %v", err)
	}
}

func TestAssembleAcceptsGroupedEmptyResult(t *testing.T) {
	page, err := Assemble(fixtureAggregateStatement(t), [][]rowdecode.Value{})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := page.Groups()
	if err != nil || len(groups) != 0 || groups == nil {
		t.Fatalf("groups = %#v, %v", groups, err)
	}
}

func fixtureAggregateRow(
	t testing.TB, statement compiler.AggregateStatement,
	group, count, revenue any,
) []rowdecode.Value {
	t.Helper()
	metadata, err := loadMetadata(statement)
	if err != nil {
		t.Fatal(err)
	}
	driver := []any{group, count, revenue}
	columns := append(append([]column{}, metadata.groups...), metadata.measures...)
	row := make([]rowdecode.Value, len(columns))
	for index := range columns {
		row[index], err = rowdecode.Decode(columns[index].fieldType, driver[index])
		if err != nil {
			t.Fatal(err)
		}
	}
	return row
}
