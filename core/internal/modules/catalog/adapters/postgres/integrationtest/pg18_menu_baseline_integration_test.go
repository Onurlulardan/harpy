//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	catalogmodule "github.com/Onurlulardan/harpy/core/internal/modules/catalog"
)

func assertCatalogMenuBaselineState(t *testing.T, ctx context.Context,
	database *sql.DB, schema string, want int,
) {
	t.Helper()
	var definitions, versions, items int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM `+schema+`.catalog_menu_definitions
 WHERE definition_id='018f6000-0000-7000-8000-000000000001'),
(SELECT count(*) FROM `+schema+`.catalog_menu_versions
 WHERE version_id='018f6000-0000-7000-8000-000000000002'),
(SELECT count(*) FROM `+schema+`.catalog_menu_items
 WHERE version_id='018f6000-0000-7000-8000-000000000002')`).
		Scan(&definitions, &versions, &items)
	if err != nil || definitions != want || versions != want || items != 0 {
		t.Fatalf("Catalog %s empty-menu baseline=%d/%d/%d want=%d/%d/0 err=%v",
			schema, definitions, versions, items, want, want, err)
	}
}

func assertCatalogMenuRepairRejectsTamper(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const schema = "catalog_b"
	_, err := database.ExecContext(ctx, `INSERT INTO `+schema+`.catalog_menu_definitions(
definition_id,revision,active_version_id,updated_at)
VALUES('018f6000-0000-7000-8000-000000000001',1,NULL,clock_timestamp());
INSERT INTO `+schema+`.catalog_menu_versions(
version_id,definition_id,version_number,content_sha256,created_at)
VALUES('018f6000-0000-7000-8000-000000000002',
'018f6000-0000-7000-8000-000000000001',1,
'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
clock_timestamp());
UPDATE `+schema+`.catalog_menu_definitions
SET active_version_id='018f6000-0000-7000-8000-000000000002'
WHERE definition_id='018f6000-0000-7000-8000-000000000001'`)
	if err != nil {
		t.Fatal(err)
	}
	defer clearCatalogMenuTamper(t, ctx, database, schema)

	err = executeCatalogMenuRepair(t, ctx, database, schema, nil)
	if err == nil || !strings.Contains(err.Error(), "not the exact baseline") {
		t.Fatalf("Catalog modified baseline repair denial=%v", err)
	}
	assertCatalogMenuBaselineState(t, ctx, database, schema, 1)
}

func assertCatalogMenuRepairRejectsAuthorityMismatch(t *testing.T,
	ctx context.Context, database *sql.DB,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `UPDATE platform.domain_cutovers
SET generation=generation+1 WHERE scope_kind='tenant' AND scope_id=$1
AND domain='catalog'`, catalogTenantB); err != nil {
		t.Fatal(err)
	}
	err = executeCatalogMenuRepair(t, ctx, database, "catalog_b", transaction)
	if err == nil || !strings.Contains(err.Error(), "authority differs") {
		t.Fatalf("Catalog mismatched authority repair denial=%v", err)
	}
}

func executeCatalogMenuRepair(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, transaction *sql.Tx,
) error {
	t.Helper()
	body := catalogMenuRepairBody(t)
	owned := transaction == nil
	if owned {
		var err error
		transaction, err = database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = transaction.Rollback() }()
	}
	if _, err := transaction.ExecContext(ctx,
		"SELECT set_config('search_path',$1,true)",
		schema+",platform,pg_catalog"); err != nil {
		return err
	}
	_, err := transaction.ExecContext(ctx, body)
	return err
}

func catalogMenuRepairBody(t *testing.T) string {
	t.Helper()
	documents, err := catalogmodule.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	for _, document := range documents {
		if strings.Contains(document.Path(), "20260728105100_") &&
			strings.Contains(document.Path(), ".v1_adopted.sql") {
			return string(document.Content())
		}
	}
	t.Fatal("Catalog menu target repair migration is missing")
	return ""
}

func clearCatalogMenuTamper(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE `+schema+`.catalog_menu_definitions
SET active_version_id=NULL
WHERE definition_id='018f6000-0000-7000-8000-000000000001';
DELETE FROM `+schema+`.catalog_menu_versions
WHERE version_id='018f6000-0000-7000-8000-000000000002';
DELETE FROM `+schema+`.catalog_menu_definitions
WHERE definition_id='018f6000-0000-7000-8000-000000000001'`)
	if err != nil {
		t.Fatal(err)
	}
}
