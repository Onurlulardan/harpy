//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func corruptLayoutRepairSourceSHA(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
ALTER TABLE layout_b.system_resource_layout_publications
DISABLE TRIGGER system_resource_layout_publications_write_guard;
UPDATE layout_b.system_resource_layout_publications
SET source_definition_sha256=pg_catalog.sha256(
    pg_catalog.convert_to('corrupt','UTF8'))
WHERE resource_kind='organization';
ALTER TABLE layout_b.system_resource_layout_publications
ENABLE ALWAYS TRIGGER system_resource_layout_publications_write_guard;`)
	if err != nil {
		t.Fatal(err)
	}
}

func removeLayoutRepairPublication(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
ALTER TABLE layout_missing.system_resource_layout_publications
DISABLE TRIGGER system_resource_layout_publications_delete_guard;
DELETE FROM layout_missing.system_resource_layout_publications
WHERE resource_kind='role';
ALTER TABLE layout_missing.system_resource_layout_publications
ENABLE ALWAYS TRIGGER system_resource_layout_publications_delete_guard;`)
	if err != nil {
		t.Fatal(err)
	}
}

func corruptLayoutRepairSemanticCatalog(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
ALTER TABLE layout_semantic.system_resource_layout_publications
DISABLE TRIGGER system_resource_layout_publications_write_guard;
UPDATE layout_semantic.system_resource_layout_publications
SET catalog_json=pg_catalog.jsonb_set(
    catalog_json::jsonb,'{fields,1,type_id}','"uuid"'::jsonb)::text,
    catalog_sha256=pg_catalog.sha256(pg_catalog.convert_to(
      pg_catalog.jsonb_set(catalog_json::jsonb,
        '{fields,1,type_id}','"uuid"'::jsonb)::text,'UTF8'))
WHERE resource_kind='user';
ALTER TABLE layout_semantic.system_resource_layout_publications
ENABLE ALWAYS TRIGGER system_resource_layout_publications_write_guard;
ALTER TABLE layout_semantic.system_resource_layout_installation_baselines
DISABLE TRIGGER system_resource_layout_installation_baselines_write_guard;
UPDATE layout_semantic.system_resource_layout_installation_baselines b
SET baseline_sha256=pg_catalog.sha256(
 pg_catalog.uuid_send(b.operation_id)||pg_catalog.uuid_send(b.tenant_id)||
 pg_catalog.uuid_send(b.user_revision_id)||u.catalog_sha256||u.presentation_sha256||
 pg_catalog.uuid_send(b.organization_revision_id)||o.catalog_sha256||
 o.presentation_sha256||pg_catalog.uuid_send(b.role_revision_id)||
 r.catalog_sha256||r.presentation_sha256)
FROM layout_semantic.system_resource_layout_publications u,
     layout_semantic.system_resource_layout_publications o,
     layout_semantic.system_resource_layout_publications r
WHERE u.resource_kind='user' AND o.resource_kind='organization'
  AND r.resource_kind='role';
ALTER TABLE layout_semantic.system_resource_layout_installation_baselines
ENABLE ALWAYS TRIGGER system_resource_layout_installation_baselines_write_guard;`)
	if err != nil {
		t.Fatal(err)
	}
}

func applyLayoutMigrationExpectErrorPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema, name string,
) error {
	t.Helper()
	connection, transaction := beginLayoutOwnerTx(t, ctx, database, schema)
	defer connection.Close()
	_, err := transaction.ExecContext(ctx, string(readLayoutMigration(
		t, name,
	)))
	if err != nil {
		_ = transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func callInstallationLayoutRepairPG18(
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
system_resource_layout_repair_installation_defaults_v2(
'curated-installation-default-successor.v1')`).Scan(&changed)
	}
	if err != nil {
		_ = transaction.Rollback()
		return 0, err
	}
	return changed, transaction.Commit()
}

func assertLayoutRepairStatePG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema, generations string, revisions int,
) {
	t.Helper()
	var got string
	err := database.QueryRowContext(ctx, `SELECT pg_catalog.string_agg(
resource_kind||':'||generation::text,',' ORDER BY resource_kind)
FROM `+schema+`.system_resource_layout_publications`).Scan(&got)
	if err != nil || got != generations {
		t.Fatalf("layout generations=%q want=%q: %v", got, generations, err)
	}
	var count int
	err = database.QueryRowContext(ctx, `SELECT count(*) FROM `+schema+
		`.system_resource_layout_revisions`).Scan(&count)
	if err != nil || count != revisions {
		t.Fatalf("layout revisions=%d want=%d: %v", count, revisions, err)
	}
}

func layoutRepairSnapshotPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema, source string,
) string {
	t.Helper()
	query := `SELECT pg_catalog.string_agg(value,E'\n' ORDER BY value) FROM (`
	if source == "publications" {
		query += `SELECT resource_kind||catalog_json||
pg_catalog.encode(catalog_sha256,'hex') AS value FROM ` + schema +
			`.system_resource_layout_publications`
	} else {
		query += `SELECT id::text||resource_kind||version::text||state||
definition_json||pg_catalog.encode(definition_sha256,'hex')||
created_at::text||updated_at::text||COALESCE(published_at::text,'')||
row_revision::text AS value FROM ` + schema +
			`.system_resource_layout_revisions WHERE id IN ('` +
			layoutContractUserRevision + "','" +
			layoutContractOrganizationRevision + "','" +
			layoutContractRoleRevision + "')"
	}
	query += `) snapshot`
	var value string
	if err := database.QueryRowContext(ctx, query).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
