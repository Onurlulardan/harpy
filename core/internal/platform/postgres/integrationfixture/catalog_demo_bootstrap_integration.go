//go:build integration

package integrationfixture

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func (environment *CatalogDemoEnvironment) prepareBootstrap(
	t *testing.T, ctx context.Context, origin string,
) {
	t.Helper()
	environment.execFixture(t, ctx, "bootstrap_tables.fixture")
	environment.execFixture(t, ctx, "extensions_fresh.fixture")
	migrationDigest, err := schemacontract.IntegrationMigrationManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	commands := []struct {
		name string
		args []any
	}{
		{"seed_environment.fixture", []any{origin}},
		{"seed_migration_run.fixture", []any{origin, migrationDigest}},
		{"seed_schema_contract.fixture", []any{migrationDigest}},
		{"seed_bootstrap_receipts.fixture", []any{origin}},
		{"seed_release_generation.fixture",
			[]any{origin, migrationDigest, runtimev1.PostgresImageDigest}},
	}
	extensionSchema, extensionOwner := "platform_ext", "harpy_extension_admin"
	commands = append(commands, struct {
		name string
		args []any
	}{"seed_extension_registry.fixture",
		[]any{origin, extensionSchema, extensionOwner}})
	for _, command := range commands {
		if _, err = environment.owner.ExecContext(
			ctx, string(catalogDemoFixture(command.name)), command.args...,
		); err != nil {
			t.Fatalf("execute Catalog demo bootstrap %s: %v", command.name, err)
		}
	}
}

func applyCatalogDemoOutbox(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	documents []migration.Document,
	origin string,
) {
	t.Helper()
	marker := "/migrations/" + origin + "/20260715020300_"
	count := 0
	for _, document := range documents {
		if !strings.Contains(filepath.ToSlash(document.Path()), marker) {
			continue
		}
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply outbox %s: %v", document.Path(), err)
		}
		count++
	}
	if count != 1 {
		t.Fatalf("outbox migration count=%d", count)
	}
	if _, err := database.ExecContext(ctx,
		string(catalogDemoFixture("outbox_grant.fixture"))); err != nil {
		t.Fatal(err)
	}
}
