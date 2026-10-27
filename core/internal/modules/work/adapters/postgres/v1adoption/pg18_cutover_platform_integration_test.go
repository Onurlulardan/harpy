//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

var workCutoverPlatformMigrations = []string{
	"20260713020000_create_domain_cutovers.fresh.sql",
	"20260713020100_create_cutover_groups.fresh.sql",
	"20260713020200_create_cutover_group_members.fresh.sql",
	"20260713020300_create_cutover_operations.fresh.sql",
	"20260713020400_create_cutover_operation_relations.fresh.sql",
	"20260713020450_create_cutover_operation_checkpoint_keys.fresh.sql",
	"20260713020500_create_cutover_operation_columns.fresh.sql",
	"20260713020600_create_cutover_operation_procedures.fresh.sql",
	"20260713020700_create_cutover_copy_checkpoints.fresh.sql",
	"20260713020800_create_cutover_initial_copy_receipts.fresh.sql",
	"20260713020900_create_applied_change_receipts.fresh.sql",
	"20260713021000_create_cutover_change_journal.fresh.sql",
	"20260713021100_create_cutover_credential_receipts.fresh.sql",
	"20260713021200_create_cutover_transition_events.fresh.sql",
	"20260713021300_create_contract_retirements.fresh.sql",
	"20260713021400_create_cutover_reject_delete.fresh.sql",
	"20260713021500_create_cutover_reject_mutation.fresh.sql",
	"20260714012000_add_procedure_contract_id.fresh.sql",
	"20260714012100_create_cutover_batch_receipts.fresh.sql",
	"20260714012500_create_cutover_prepare_batch_v1.fresh.sql",
	"20260714012600_create_cutover_record_batch_v1.fresh.sql",
}

func prepareWorkCutoverPlatform(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOBYPASSRLS;
CREATE ROLE harpy_adopter LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1
PASSWORD 'work-adopter-session';
CREATE ROLE harpy_cutover_decoder NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0 PASSWORD NULL;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;`)
	if err != nil {
		t.Fatalf("prepare Work cutover catalog: %v", err)
	}
	directory := filepath.Join("..", "..", "..", "..", "..", "platform",
		"cutover", "migrations", "fresh")
	for _, name := range workCutoverPlatformMigrations {
		document, readErr := os.ReadFile(filepath.Join(directory, name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, err = database.ExecContext(ctx, string(document)); err != nil {
			t.Fatalf("apply Work cutover dependency %s: %v", name, err)
		}
	}
	if _, err = database.ExecContext(ctx,
		"GRANT USAGE ON SCHEMA platform TO harpy_adopter"); err != nil {
		t.Fatal(err)
	}
}

func prepareWorkTenantIdentity(t *testing.T, ctx context.Context, database *sql.DB,
	schema, identityOrigin string,
) {
	t.Helper()
	directory := filepath.Join("..", "..", "..", "..", "tenancy", "migrations",
		"tenant", "fresh")
	name := "20260714011000_create_tenancy_schema_identity.fresh.sql"
	document, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	if len(document) == 0 {
		t.Fatal("fresh tenancy identity migration is empty")
	}
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		if _, execErr := transaction.ExecContext(ctx, string(document)); execErr != nil {
			t.Fatalf("apply tenant identity to %s: %v", schema, execErr)
		}
		_, execErr := transaction.ExecContext(ctx, `
INSERT INTO tenancy_schema_identity(singleton,tenant_id,schema_name,origin,created_at)
VALUES (true,$1,$2,$3,'2026-07-19T00:00:00Z')`,
			workTenantID(schema), schema, identityOrigin)
		if execErr != nil {
			t.Fatalf("seed tenant identity for %s: %v", schema, execErr)
		}
	})
}

func workTenantID(schema string) string {
	identities := map[string]string{
		"work_fresh_a":   "21000000-0000-4000-8000-000000000001",
		"work_fresh_b":   "21000000-0000-4000-8000-000000000002",
		"work_adopted_a": "21000000-0000-4000-8000-000000000003",
		"work_adopted_b": "21000000-0000-4000-8000-000000000004",
	}
	return identities[schema]
}
