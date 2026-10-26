//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertCuratedParityRepairPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	assertLayoutRepairStatePG18(
		t, ctx, database, layoutContractSchema,
		"organization:3,role:3,user:3", 9,
	)
	for _, expected := range []struct {
		resource, surface, field string
		present                  bool
	}{
		{"user", "record_create", "a1000000-0000-4000-8000-000000000003", false},
		{"role", "record_create", "b1000000-0000-4000-8000-000000000009", true},
		{"organization", "record_create", "c1000000-0000-4000-8000-00000000000c", true},
		{"organization", "record_update", "c1000000-0000-4000-8000-00000000000c", true},
	} {
		if got := curatedParityFieldExistsPG18(
			t, ctx, database, layoutContractSchema,
			expected.resource, expected.surface, expected.field,
		); got != expected.present {
			t.Fatalf("%s %s field %s present=%v want=%v",
				expected.resource, expected.surface, expected.field,
				got, expected.present)
		}
	}
	var ownerMode string
	err := database.QueryRowContext(ctx, `
SELECT field->>'mode'
FROM `+layoutContractSchema+`.system_resource_layout_publications publication
CROSS JOIN LATERAL pg_catalog.jsonb_array_elements(
 publication.catalog_json::jsonb->'fields') field
WHERE publication.resource_kind='organization'
  AND field->>'id'='c1000000-0000-4000-8000-00000000000c'`).Scan(&ownerMode)
	if err != nil || ownerMode != "editable" {
		t.Fatalf("organization owner catalog mode=%q: %v", ownerMode, err)
	}
	for _, expected := range []struct{ surface, label string }{
		{"record_create", "Oluştur"},
		{"record_update", "Güncelle"},
		{"record_detail", "Ayrıntılar"},
	} {
		var label string
		err = database.QueryRowContext(ctx, `
SELECT translation->>'label'
FROM `+layoutContractSchema+`.system_resource_layout_publications publication
JOIN `+layoutContractSchema+`.system_resource_layout_revisions revision
  ON revision.id=publication.revision_id
CROSS JOIN LATERAL pg_catalog.jsonb_array_elements(
  revision.definition_json::jsonb->'surfaces') surface
CROSS JOIN LATERAL pg_catalog.jsonb_array_elements(
  surface->'translations') translation
WHERE publication.resource_kind='organization'
  AND surface->>'kind'=$1 AND translation->>'locale'='tr'`,
			expected.surface).Scan(&label)
		if err != nil || label != expected.label {
			t.Fatalf("%s Turkish label=%q want=%q: %v",
				expected.surface, label, expected.label, err)
		}
	}
	changed, err := callCuratedParityRepairPG18(
		ctx, database, layoutContractSchema,
	)
	if err != nil || changed != 0 {
		t.Fatalf("curated parity replay changed=%d: %v", changed, err)
	}
}

func assertMixedCustomCuratedParityPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	assertLayoutRepairStatePG18(
		t, ctx, database, "layout_a",
		"organization:3,role:3,user:2", 8,
	)
	var revision string
	err := database.QueryRowContext(ctx, `
SELECT revision_id::text FROM layout_a.system_resource_layout_publications
WHERE resource_kind='user'`).Scan(&revision)
	if err != nil || revision != layoutSecondRevision {
		t.Fatalf("custom user successor revision=%q: %v", revision, err)
	}
}

func curatedParityFieldExistsPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema, resource, surface, fieldID string,
) bool {
	t.Helper()
	var found bool
	err := database.QueryRowContext(ctx, `
SELECT EXISTS (
 SELECT 1
 FROM `+schema+`.system_resource_layout_publications publication
 CROSS JOIN LATERAL pg_catalog.jsonb_array_elements(
  (SELECT revision.definition_json::jsonb->'surfaces'
   FROM `+schema+`.system_resource_layout_revisions revision
   WHERE revision.id=publication.revision_id)) surface
 CROSS JOIN LATERAL pg_catalog.jsonb_array_elements(surface->'sections') section
 CROSS JOIN LATERAL pg_catalog.jsonb_array_elements(section->'fields') field
 WHERE publication.resource_kind=$1 AND surface->>'kind'=$2
   AND field->>'field_id'=$3
)`, resource, surface, fieldID).Scan(&found)
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func callCuratedParityRepairPG18(
	ctx context.Context, database *sql.DB, schema string,
) (int, error) {
	connection, err := database.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	var bound string
	err = transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		schema+", pg_catalog").Scan(&bound)
	var changed int
	if err == nil {
		err = transaction.QueryRowContext(ctx, `SELECT
system_resource_layout_repair_curated_defaults_v3(
'curated-system-resource-parity-successor.v2')`).Scan(&changed)
	}
	if err != nil {
		_ = transaction.Rollback()
		return 0, err
	}
	return changed, transaction.Commit()
}

func corruptCuratedParitySourceSHA(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
ALTER TABLE layout_parity_corrupt.system_resource_layout_publications
DISABLE TRIGGER system_resource_layout_publications_write_guard;
UPDATE layout_parity_corrupt.system_resource_layout_publications
SET source_definition_sha256=pg_catalog.sha256(
    pg_catalog.convert_to('corrupt','UTF8'))
WHERE resource_kind='user';
ALTER TABLE layout_parity_corrupt.system_resource_layout_publications
ENABLE ALWAYS TRIGGER system_resource_layout_publications_write_guard;`)
	if err != nil {
		t.Fatal(err)
	}
}
