package compiler

import (
	"crypto/sha256"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompileRecordDeterministicGolden(t *testing.T) {
	name := fixtureField(t,
		"11111111-1111-4111-8111-111111111111", "display_name", "fld_name",
		fixtureFoldedTextType(t, false), "fld_name_folded", "",
		[]api.Operator{api.OperatorContains}, false, true,
	)
	owner := fixtureField(t,
		"22222222-2222-4222-8222-222222222222", "owner_id", "fld_owner",
		fixtureUUIDType(t), "", "", []api.Operator{api.OperatorEq}, false, false,
	)
	status := fixtureField(t,
		"33333333-3333-4333-8333-333333333333", "status", "fld_status",
		fixtureOrderedEnumType(t), "", "", []api.Operator{api.OperatorGTE}, true, false,
	)
	stable := fixtureField(t,
		"44444444-4444-4444-8444-444444444444", "id", "fld_id",
		fixtureUUIDType(t), "", "", []api.Operator{}, true, false,
	)
	ownerValue := fixtureScalar(t, api.LogicalTypeUUID,
		"55555555-5555-4555-8555-555555555555")
	mandatoryPredicate, err := resolved.NewMandatoryPredicate(
		owner, api.OperatorEq, []resolved.Value{ownerValue},
	)
	if err != nil {
		t.Fatal(err)
	}
	mandatory, err := resolved.NewMandatoryPredicateFilter(mandatoryPredicate)
	if err != nil {
		t.Fatal(err)
	}
	visibility, err := resolved.NewPredicateResourceVisibility(
		sha256.Sum256([]byte("requirement")), fixtureReference(t, "forms"),
		fixturePath(t, ""), mandatory,
	)
	if err != nil {
		t.Fatal(err)
	}
	whereValue := fixtureScalar(t, api.LogicalTypeEnum, "CLOSED")
	where := fixturePredicateFilter(
		t, status, api.OperatorGTE, []resolved.Value{whereValue},
	)
	searchValue := fixtureScalar(t, api.LogicalTypeText, "needle")
	search, err := resolved.NewSearchTerm(name, searchValue, fixturePath(t, "/search"))
	if err != nil {
		t.Fatal(err)
	}
	query := fixtureRecordQuery(t,
		[]resolved.Projection{
			fixtureProjection(t, name, "display", false),
			fixtureProjection(t, owner, "owner", false),
		},
		&where,
		[]resolved.Sort{
			fixtureSort(t, status, api.DirectionDescending, false),
			fixtureSort(t, stable, api.DirectionAscending, true),
		},
		[]resolved.SearchTerm{search}, 2,
	)
	statement, err := compileRecord(fixtureRecordInput(t, "forms", query, visibility))
	if err != nil {
		t.Fatal(err)
	}
	wantSQL := `SELECT "fld_name", "fld_owner", "fld_status", "fld_id" FROM "frm_records" WHERE ("fld_owner" = $1) AND (pg_catalog.array_position(CAST($2 AS pg_catalog.text[]), "fld_status") >= $3) AND (("fld_name_folded" COLLATE pg_catalog.pg_unicode_fast LIKE $4 ESCAPE E'\\')) ORDER BY pg_catalog.array_position(CAST($5 AS pg_catalog.text[]), "fld_status") DESC NULLS LAST, "fld_id" ASC NULLS LAST LIMIT $6`
	actualSQL, sqlErr := statement.SQL()
	if sqlErr != nil || actualSQL != wantSQL {
		t.Fatalf("SQL mismatch\nwant: %s\n got: %s\nerr: %v", wantSQL, actualSQL, sqlErr)
	}
	wantArguments := []any{
		"55555555-5555-4555-8555-555555555555",
		[]string{"OPEN", "CLOSED"}, 20, "%needle%",
		[]string{"OPEN", "CLOSED"}, 3,
	}
	arguments, err := statement.Arguments()
	if err != nil || !reflect.DeepEqual(arguments, wantArguments) {
		t.Fatalf("arguments mismatch: %#v err=%v", arguments, err)
	}
	assertGoldenColumns(t, statement)
}

func assertGoldenColumns(t *testing.T, statement RecordStatement) {
	t.Helper()
	columns, err := statement.Columns()
	if err != nil || len(columns) != 4 {
		t.Fatalf("columns mismatch: count=%d err=%v", len(columns), err)
	}
	for index, expected := range []struct {
		output    string
		sort      int
		direction api.Direction
		bound     int
	}{
		{"display", -1, 0, 0}, {"owner", -1, 0, 0},
		{"", 0, api.DirectionDescending, 256},
		{"", 1, api.DirectionAscending, 256},
	} {
		output, present, outputErr := columns[index].OutputKey()
		sortOrdinal, sorted, sortErr := columns[index].SortOrdinal()
		direction, bound, metadataPresent, metadataErr := columns[index].SortMetadata()
		if outputErr != nil || sortErr != nil || output != expected.output ||
			present != (expected.output != "") || sortOrdinal != expected.sort ||
			sorted != (expected.sort >= 0) || metadataErr != nil ||
			metadataPresent != sorted || direction != expected.direction || bound != expected.bound {
			t.Fatalf("column %d metadata mismatch", index)
		}
		if fieldType, typeErr := columns[index].FieldType(); typeErr != nil ||
			fieldType.Validate() != nil {
			t.Fatalf("column %d exact field type missing", index)
		}
	}
}
