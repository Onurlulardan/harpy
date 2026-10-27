//go:build integration

package integrationfixture

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

var catalogDemoCutoverMigrations = []string{
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

func (environment *CatalogDemoEnvironment) prepare(
	t *testing.T,
	ctx context.Context,
	root string,
	origin string,
	catalogDocuments []migration.Document,
	formsDocuments []migration.Document,
	analyticsDocuments []migration.Document,
	asyncDocuments []migration.Document,
) {
	t.Helper()
	environment.execFixture(t, ctx, "base.fixture")
	for _, name := range catalogDemoCutoverMigrations {
		path := filepath.Join(root, "core", "internal", "platform", "cutover",
			"migrations", "fresh", name)
		document, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = environment.owner.ExecContext(ctx, string(document)); err != nil {
			t.Fatalf("apply Catalog demo cutover dependency %s: %v", name, err)
		}
	}
	environment.execFixture(t, ctx, "authority_fresh.fixture")
	environment.prepareBootstrap(t, ctx, origin)
	environment.applyRuntimeFoundation(t, ctx, root, origin)
	platformCount := applyCatalogDemoDocuments(t, ctx, environment.owner,
		catalogDocuments, "/migrations/platform/"+origin+"/", false)
	tenantCount := applyCatalogDemoDocuments(t, ctx, environment.owner,
		catalogDocuments, "/migrations/tenant/"+origin+"/", true)
	if platformCount < 6 || tenantCount < 35 {
		t.Fatalf("Catalog migration counts platform=%d tenant=%d",
			platformCount, tenantCount)
	}
	applyCatalogDemoOutbox(t, ctx, environment.owner, asyncDocuments, origin)
	environment.execFixture(t, ctx, "schema_jobs.fixture")
	formsCount := applyCatalogDemoDocuments(t, ctx, environment.owner,
		formsDocuments, "/migrations/tenant/"+origin+"/", true)
	if formsCount < 80 {
		t.Fatalf("Forms tenant migration count=%d", formsCount)
	}
	if len(analyticsDocuments) > 0 {
		environment.applyCatalogDemoAnalytics(
			t, ctx, analyticsDocuments, origin,
		)
	}
}

func (environment *CatalogDemoEnvironment) execFixture(
	t *testing.T, ctx context.Context, name string,
) {
	t.Helper()
	if _, err := environment.owner.ExecContext(
		ctx, string(catalogDemoFixture(name)),
	); err != nil {
		t.Fatalf("execute Catalog demo fixture %s: %v", name, err)
	}
}

func (environment *CatalogDemoEnvironment) applyRuntimeFoundation(
	t *testing.T, ctx context.Context, root, origin string,
) {
	t.Helper()
	paths := []string{
		filepath.Join(root, "core", "internal", "modules", "tenancy",
			"migrations", "platform", origin,
			"20260718050600_create_tenancy_runtime_resolver_v1."+origin+".sql"),
		filepath.Join(root, "core", "internal", "platform", "cutover",
			"migrations", origin,
			"20260718050601_create_cutover_runtime_locks_v1."+origin+".sql"),
	}
	for _, path := range paths {
		document, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = environment.owner.ExecContext(ctx, string(document)); err != nil {
			t.Fatalf("apply runtime foundation %s: %v", path, err)
		}
	}
}

func applyCatalogDemoDocuments(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	documents []migration.Document,
	needle string,
	tenant bool,
) int {
	t.Helper()
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].Path() < documents[right].Path()
	})
	count := 0
	for _, document := range documents {
		if !strings.Contains(filepath.ToSlash(document.Path()), needle) {
			continue
		}
		transaction, err := database.BeginTx(ctx, nil)
		if err == nil && tenant {
			_, err = transaction.ExecContext(ctx,
				string(catalogDemoFixture("set_search_path.fixture")))
		}
		if err == nil {
			_, err = transaction.ExecContext(ctx, string(document.Content()))
		}
		if err == nil {
			err = transaction.Commit()
		} else {
			_ = transaction.Rollback()
		}
		if err != nil {
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
		count++
	}
	return count
}
