//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
)

var billingCutoverPlatformMigrations = map[string]bool{
	"20260713020000_create_domain_cutovers":                   true,
	"20260713020100_create_cutover_groups":                    true,
	"20260713020200_create_cutover_group_members":             true,
	"20260713020300_create_cutover_operations":                true,
	"20260713020400_create_cutover_operation_relations":       true,
	"20260713020450_create_cutover_operation_checkpoint_keys": true,
	"20260713020500_create_cutover_operation_columns":         true,
	"20260713020600_create_cutover_operation_procedures":      true,
	"20260713020700_create_cutover_copy_checkpoints":          true,
	"20260713020800_create_cutover_initial_copy_receipts":     true,
	"20260713020900_create_applied_change_receipts":           true,
	"20260713021000_create_cutover_change_journal":            true,
	"20260714012000_add_procedure_contract_id":                true,
	"20260714012100_create_cutover_batch_receipts":            true,
	"20260714012500_create_cutover_prepare_batch_v1":          true,
	"20260714012600_create_cutover_record_batch_v1":           true,
}

func applyBillingCutoverPlatformMigrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents, err := platformcutover.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(documents, func(left, right int) bool {
		return filepath.Base(documents[left].Path()) <
			filepath.Base(documents[right].Path())
	})
	applied := 0
	suffix := "." + origin + ".sql"
	for _, document := range documents {
		name := filepath.Base(document.Path())
		if !strings.HasSuffix(name, suffix) ||
			!billingCutoverPlatformMigrations[strings.TrimSuffix(name, suffix)] {
			continue
		}
		if _, err = database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply Billing cutover dependency %s: %v", name, err)
		}
		applied++
	}
	if applied != len(billingCutoverPlatformMigrations) {
		t.Fatalf("Billing cutover dependencies applied=%d want=%d",
			applied, len(billingCutoverPlatformMigrations))
	}
}

func prepareBillingCutoverSource(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `
CREATE TABLE platform.billing_v1_adoption_source(id uuid PRIMARY KEY);
GRANT USAGE ON SCHEMA platform TO harpy_adopter;`); err != nil {
		t.Fatal(err)
	}
}
