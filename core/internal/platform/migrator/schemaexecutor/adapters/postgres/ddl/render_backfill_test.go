package ddl

import (
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestRenderBackfillUsesExactFieldGenerations(t *testing.T) {
	const fieldID = "00000000-0000-4000-8000-000000000006"
	mapping := schemav1.IdentifierMapping{
		Form: schemav1.Identifier{PhysicalName: "frm_example"},
		Fields: []schemav1.FieldIdentifier{
			{LogicalID: fieldID, Generation: 7, PhysicalName: "fld_example_g7"},
			{LogicalID: fieldID, Generation: 8, PhysicalName: "fld_example_g8"},
		},
	}
	arguments := schemav1.BackfillArguments{
		SourceFieldID: fieldID, SourceGeneration: 7,
		TargetFieldID: fieldID, TargetGeneration: 8,
		Converter: schemav1.ConvertIdentity,
	}
	statement, err := renderBackfillStatement("tenant_example", mapping, arguments)
	if err != nil {
		t.Fatal(err)
	}
	expected := "WITH selected AS MATERIALIZED (SELECT id, row_version, " +
		"\"fld_example_g7\" AS source_value FROM \"tenant_example\".\"frm_example\" " +
		"WHERE ($1::uuid IS NULL OR id > $1::uuid) ORDER BY id LIMIT 1000 FOR UPDATE), " +
		"updated AS (UPDATE \"tenant_example\".\"frm_example\" AS target SET " +
		"\"fld_example_g8\" = selected.source_value, " +
		"updated_at = pg_catalog.statement_timestamp(), row_version = target.row_version + 1 " +
		"FROM selected WHERE target.id = selected.id AND target.row_version = selected.row_version " +
		"RETURNING target.id) SELECT (SELECT id::text FROM selected ORDER BY id DESC LIMIT 1), " +
		"(SELECT pg_catalog.count(*)::bigint FROM selected), " +
		"(SELECT id::text FROM updated ORDER BY id DESC LIMIT 1), " +
		"(SELECT pg_catalog.count(*)::bigint FROM updated)"
	if statement.Text != expected {
		t.Fatalf("backfill SQL mismatch\nwant: %s\n got: %s", expected, statement.Text)
	}
}
