//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

var analyticsCutoverPlatformMigrations = []string{
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

func applyAnalyticsCutoverDependencies(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `CREATE ROLE harpy_cutover_decoder
NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join("..", "..", "..", "..", "..", "platform",
		"cutover", "migrations", "fresh")
	for _, name := range analyticsCutoverPlatformMigrations {
		document, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = database.ExecContext(ctx, string(document)); err != nil {
			t.Fatalf("apply Analytics cutover dependency %s: %v", name, err)
		}
	}
}
