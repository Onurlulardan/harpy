//go:build integration

package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

var workflowDependencyTables = []string{
	"cutover_operations", "cutover_operation_relations",
	"cutover_operation_checkpoint_keys", "cutover_operation_columns",
	"cutover_operation_procedures", "cutover_initial_copy_receipts",
	"cutover_batch_receipts", "cutover_copy_checkpoints", "applied_change_receipts",
	"cutover_change_journal", "cutover_credential_receipts",
	"tenant_migration_requirement_catalogs", "tenant_migration_requirements",
	"tenant_migration_requirement_targets", "tenant_migration_checkpoints", "tenants",
	"release_generations", "schema_contracts",
}

func prepareWorkflowSchemas(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	if _, err := admin.Exec(ctx, `CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE SCHEMA cutover_authority AUTHORIZATION harpy_owner;
SET ROLE harpy_owner;`); err != nil {
		t.Fatal(err)
	}
	for _, name := range workflowDependencyTables {
		statement := workflowDependencyTableDDL(name)
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("create dependency %s: %v", name, err)
		}
	}
	prepareWorkflowAuthorityTables(t, ctx, admin)
	prepareWorkflowFingerprintFunctions(t, ctx, admin)
	prepareWorkflowAuthorityFunctions(t, ctx, admin)
	if _, err := admin.Exec(ctx, "RESET ROLE"); err != nil {
		t.Fatal(err)
	}
}

func workflowDependencyTableDDL(name string) string {
	if name == "cutover_credential_receipts" {
		return `CREATE TABLE platform.cutover_credential_receipts (
 operation_id uuid NOT NULL, run_id uuid NOT NULL,
 seal_receipt_sha256 text NOT NULL, seal_receipt_document bytea NOT NULL,
 cleanup_receipt_sha256 text NOT NULL, cleanup_receipt_document bytea NOT NULL,
 PRIMARY KEY(operation_id,run_id));
GRANT USAGE ON SCHEMA platform TO harpy_cutover_authority_owner;
GRANT SELECT,INSERT ON platform.cutover_credential_receipts
TO harpy_cutover_authority_owner`
	}
	return "CREATE TABLE platform." + pgx.Identifier{name}.Sanitize() +
		" (fixture_id integer)"
}

func prepareWorkflowAuthorityTables(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	_, err := admin.Exec(ctx, `CREATE TABLE cutover_authority.authority_adoption_inventory (
 catalog_generation integer PRIMARY KEY, operation_id uuid UNIQUE NOT NULL,
 receipt_id uuid UNIQUE NOT NULL, receipt_sha256 text NOT NULL, receipt_document bytea NOT NULL,
 operation_sha256 text NOT NULL, release_id uuid NOT NULL, environment_id uuid NOT NULL,
 database_origin text NOT NULL, system_identifier numeric(20,0) NOT NULL, database_oid oid NOT NULL,
 catalog_contract_sha256 text NOT NULL, catalog_instance_sha256 text NOT NULL,
 dependency_contract_sha256 text NOT NULL, dependency_instance_sha256 text NOT NULL,
 object_count integer NOT NULL, dependency_count integer NOT NULL, adopted_at timestamptz NOT NULL);
CREATE TABLE cutover_authority.finalization_authorizations (
 operation_id uuid NOT NULL, run_id uuid NOT NULL, finalizer_lease_id uuid NOT NULL,
 payload_sha256 text NOT NULL, PRIMARY KEY(operation_id,run_id));
CREATE TABLE cutover_authority.finalization_receipts (fixture_id integer);`)
	if err != nil {
		t.Fatal(err)
	}
}

func prepareWorkflowFingerprintFunctions(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `CREATE FUNCTION cutover_authority.catalog_definition_sha256_v1(
 p_kind text,p_object_oid oid) RETURNS text LANGUAGE sql SECURITY DEFINER
 SET search_path=pg_catalog AS $$SELECT pg_catalog.encode(
 pg_catalog.sha256(pg_catalog.convert_to(p_kind||':'||p_object_oid::text,'UTF8')),'hex')$$;
CREATE FUNCTION cutover_authority.catalog_acl_sha256_v1(
 p_kind text,p_object_oid oid) RETURNS text LANGUAGE sql SECURITY DEFINER
 SET search_path=pg_catalog AS $$SELECT pg_catalog.encode(
 pg_catalog.sha256(pg_catalog.convert_to('acl:'||p_kind||':'||p_object_oid::text,'UTF8')),'hex')$$;`)
	if err != nil {
		t.Fatal(err)
	}
}

func prepareWorkflowAuthorityFunctions(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	statements := []string{workflowCatalogAssertSQL, workflowStateHashSQL,
		workflowRecordCredentialReceiptsSQL, workflowObserveAuthoritySQL,
		workflowAssertCompletedSQL,
		workflowAssertTenantSQL, workflowAssertTerminalSQL,
		workflowValidateEnvelopeSQL, workflowAuthorizeSQL,
		workflowConsumeSQL, workflowFinalizeSQL, workflowVerifySQL}
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("create protected function: %v", err)
		}
	}
	_, err := admin.Exec(ctx, `REVOKE ALL ON SCHEMA cutover_authority FROM PUBLIC;
GRANT USAGE ON SCHEMA cutover_authority TO harpy_cluster_admin;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA cutover_authority FROM PUBLIC;
GRANT EXECUTE ON FUNCTION cutover_authority.observe_finalizer_authority_v1(uuid,uuid)
TO harpy_cluster_admin;
GRANT EXECUTE ON FUNCTION cutover_authority.record_credential_receipts_v1(
uuid,uuid,text,bytea,text,bytea) TO harpy_cluster_admin;
GRANT EXECUTE ON FUNCTION cutover_authority.authorize_finalization_v1(
uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)
TO harpy_cluster_admin;`)
	if err != nil {
		t.Fatal(err)
	}
}

func workflowFunctionArguments() string {
	return strings.Join([]string{"uuid", "uuid", "uuid", "text", "bytea", "text",
		"bytea", "text", "bytea", "text", "bytea", "text", "bytea", "text",
		"text", "bytea"}, ",")
}

func quotedWorkflowTable(name string) string {
	return fmt.Sprintf("platform.%s", pgx.Identifier{name}.Sanitize())
}
