//go:build integration

package migrator

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/jackc/pgx/v5"
)

const (
	tenantRunOne = "11111111-1111-4111-8111-111111111111"
	tenantRunTwo = "22222222-2222-4222-8222-222222222222"
	tenantRunV1  = "33333333-3333-4333-8333-333333333333"
	tenantOwner  = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
)

type tenantMigrationFixture struct {
	database *sql.DB
	config   *pgx.ConnConfig
}

func newTenantMigrationFixture(
	t *testing.T,
	ctx context.Context,
) tenantMigrationFixture {
	t.Helper()
	database, config := startRunnerIntegrationDatabaseWithConfig(t, ctx)
	if _, err := database.ExecContext(
		ctx, "CREATE SCHEMA platform AUTHORIZATION harpy_owner",
	); err != nil {
		t.Fatal(err)
	}
	applyTenantConsumerFoundation(t, ctx, database)
	for _, run := range []struct{ id, origin string }{
		{tenantRunOne, "fresh"}, {tenantRunTwo, "fresh"},
		{tenantRunV1, "v1_adopted"},
	} {
		insertTenantMigrationRun(t, ctx, database, run.id, run.origin)
	}
	return tenantMigrationFixture{database: database, config: config}
}

func (fixture tenantMigrationFixture) runner(
	t *testing.T,
	origin postgresv1.DatabaseOrigin,
) *Runner {
	t.Helper()
	files, selected := tenantMigrationSources(t, origin)
	return &Runner{
		database: fixture.database,
		tenantDatabase: func() (*sql.DB, error) {
			return openOneShotTenantDatabase(fixture.config.Copy()), nil
		},
		tenantFiles: files, tenantSelected: selected,
		cleanupTimeout: requiredCleanupTimeout,
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func tenantMigrationSources(
	t *testing.T,
	origin postgresv1.DatabaseOrigin,
) (fs.FS, []postgresv1.SelectedMigration) {
	t.Helper()
	directory := filepath.Join(
		"..", "..", "modules", "tenancy", "migrations", "tenant", string(origin),
	)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	files := fstest.MapFS{}
	selected := make([]postgresv1.SelectedMigration, 0, len(entries))
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		version, err := strconv.ParseInt(entry.Name()[:14], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(content)
		files[entry.Name()] = &fstest.MapFile{Data: content}
		selected = append(selected, postgresv1.SelectedMigration{
			Owner: "tenancy", Scope: "tenant", Version: version,
			TransactionMode: "transactional", Phase: "expand",
			TargetDatabaseGeneration: 2, NextMinimumRuntimeGeneration: 1,
			Origin: origin, Path: "migrations/tenant/" + entry.Name(),
			SHA256: hex.EncodeToString(digest[:]),
		})
	}
	return files, selected
}

func applyTenantConsumerFoundation(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	patterns := []string{
		filepath.Join("migrations", "fresh", "*.sql"),
		filepath.Join("..", "..", "modules", "tenancy", "migrations", "platform", "fresh", "*.sql"),
	}
	var paths []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	sort.Slice(paths, func(left, right int) bool {
		return filepath.Base(paths[left]) < filepath.Base(paths[right])
	})
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := database.ExecContext(ctx, string(content)); err != nil {
			t.Fatalf("apply %s: %v", filepath.Base(path), err)
		}
	}
}

func insertTenantMigrationRun(
	t *testing.T, ctx context.Context, database *sql.DB, runID, origin string,
) {
	t.Helper()
	digest := strings.Repeat("a", 64)
	_, err := database.ExecContext(ctx, `INSERT INTO platform.migration_runs (
run_id,release_id,environment_id,database_origin,command,state,
migration_manifest_sha256,extension_manifest_sha256,roles_manifest_sha256,
privileges_manifest_sha256,ownership_manifest_sha256,started_at)
VALUES($1,'aaaaaaaa-1111-4111-8111-111111111111',
'aaaaaaaa-2222-4222-8222-222222222222',$2,'up','running',
$3,$3,$3,$3,$3,pg_catalog.clock_timestamp())`, runID, origin, digest)
	if err != nil {
		t.Fatal(err)
	}
}

func insertTenantTarget(
	t *testing.T, ctx context.Context, database *sql.DB, target postgresstore.TenantMigrationTarget,
) {
	t.Helper()
	if _, err := database.ExecContext(
		ctx, "CREATE SCHEMA "+pgx.Identifier{target.SchemaName}.Sanitize()+
			" AUTHORIZATION harpy_owner",
	); err != nil {
		t.Fatal(err)
	}
	_, err := database.ExecContext(ctx, `INSERT INTO platform.tenants
(id,name,slug,status,settings,owner_id,schema_name,created_at,updated_at)
VALUES($1,$2,$3,'ACTIVE','{}',$4,$5,pg_catalog.clock_timestamp(),
pg_catalog.clock_timestamp())`, target.TenantID, target.SchemaName, target.SchemaName,
		tenantOwner, target.SchemaName)
	if err != nil {
		t.Fatal(err)
	}
}
