//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func prepareDocumentsCutoverPlatform(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `CREATE SCHEMA platform;
CREATE TABLE platform.tenants (id uuid PRIMARY KEY);
CREATE FUNCTION platform.platform_async_guard_append_only_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS $$
BEGIN
 RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='append-only evidence';
END;
$$;`)
	if err != nil {
		t.Fatal(err)
	}
	for _, relative := range documentsCutoverPlatformMigrations() {
		applyDocumentsPlatformMigration(t, ctx, database, relative)
	}
	_, err = database.ExecContext(ctx, `ALTER ROLE harpy_adopter
 VALID UNTIL '2099-01-01 00:00:00+00';
ALTER ROLE harpy_cutover_decoder
 VALID UNTIL '2099-01-01 00:00:00+00';`)
	if err != nil {
		t.Fatal(err)
	}
}

func documentsCutoverPlatformMigrations() []string {
	return []string{
		"core/internal/platform/cutover/migrations/fresh/20260713020000_create_domain_cutovers.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260713020100_create_cutover_groups.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260713020200_create_cutover_group_members.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260713020300_create_cutover_operations.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260713020400_create_cutover_operation_relations.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260713020450_create_cutover_operation_checkpoint_keys.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260713020600_create_cutover_operation_procedures.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260714012000_add_procedure_contract_id.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260714012100_create_cutover_batch_receipts.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260713020700_create_cutover_copy_checkpoints.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260713020900_create_applied_change_receipts.fresh.sql",
		"core/internal/platform/async/migrations/fresh/20260715020100_create_jobs.fresh.sql",
		"core/internal/platform/async/migrations/fresh/20260715020200_create_job_attempts.fresh.sql",
		"core/internal/platform/async/migrations/fresh/20260715020500_create_idempotency_receipts.fresh.sql",
		"core/internal/platform/async/migrations/fresh/20260718230700_create_documents_cutover_job_bindings_v1.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260714012500_create_cutover_prepare_batch_v1.fresh.sql",
		"core/internal/platform/cutover/migrations/fresh/20260714012600_create_cutover_record_batch_v1.fresh.sql",
		"core/internal/platform/async/migrations/fresh/20260718230703_create_documents_cutover_enqueue_v1.fresh.sql",
	}
}

func applyDocumentsPlatformMigration(t *testing.T, ctx context.Context,
	database *sql.DB, relative string,
) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(documentsCutoverRepositoryRoot(t),
		filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = database.ExecContext(ctx, string(body)); err != nil {
		t.Fatalf("apply %s: %v", relative, err)
	}
}

func documentsCutoverRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve Documents cutover fixture")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename),
		"..", "..", "..", "..", "..", "..", ".."))
}

func prepareDocumentsCutoverIdentities(t *testing.T, ctx context.Context,
	database *sql.DB, tenantID string,
) {
	t.Helper()
	for _, fixture := range []struct{ schema, origin string }{
		{"documents_fresh_a", "fresh"}, {"documents_adopted_a", "v1_adopted"},
		{"documents_adopted_b", "v1_adopted"},
	} {
		identityTenant := tenantID
		if fixture.schema == "documents_adopted_b" {
			identityTenant = "018f0000-0000-7000-8000-000000000009"
		}
		createStatement := `CREATE TABLE ` + fixture.schema + `.tenancy_schema_identity (
 singleton boolean PRIMARY KEY,tenant_id uuid NOT NULL,
 schema_name text NOT NULL,origin text NOT NULL);`
		if _, err := database.ExecContext(ctx, createStatement); err != nil {
			t.Fatal(err)
		}
		insertStatement := `INSERT INTO ` + fixture.schema + `.tenancy_schema_identity
VALUES (true,$1,$2,$3);`
		if _, err := database.ExecContext(ctx, insertStatement, identityTenant,
			fixture.schema, fixture.origin); err != nil {
			t.Fatal(err)
		}
	}
	_, err := database.ExecContext(ctx, `GRANT USAGE ON SCHEMA
 documents_fresh_a,documents_adopted_a,documents_adopted_b
 TO harpy_adopter,harpy_cutover_finalizer;`)
	if err != nil {
		t.Fatal(err)
	}
}
