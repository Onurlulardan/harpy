//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/domain"
)

const (
	freshTenantID     = "550e8400-e29b-41d4-a716-446655440000"
	testOwnerID       = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	testEnvironmentID = "10000000-0000-4000-8000-000000000001"
)

func TestTenantRegistryMigrationsAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	resetTenantMigrationCatalog(t, ctx, database)
	t.Run("fresh", func(t *testing.T) {
		runPlatformMigrations(t, ctx, database, "fresh")
		insertFreshTenant(t, ctx, database)
		assertDeploymentPolicyStorage(t, ctx, database, "onpremise")
		assertConcurrentOnPremiseBinding(t, ctx, database, freshTenantID)
		assertRuntimePrimaryTenantTarget(
			t, ctx, database, "onpremise", "onpremise_primary",
		)
		assertRuntimeTenantResolver(t, ctx, database, freshTenantID)
		assertTenantRegistryCatalog(t, ctx, database)
		assertTenantRegistryWrites(t, ctx, database, freshTenantID)
		assertAdoptionIssuesAreAppendOnly(t, ctx, database)
	})
	resetTenantMigrationCatalog(t, ctx, database)
	t.Run("saas", func(t *testing.T) {
		runPlatformMigrations(t, ctx, database, "fresh")
		insertFreshTenant(t, ctx, database)
		assertDeploymentPolicyStorage(t, ctx, database, "saas")
		assertSaaSCustomerGuards(t, ctx, database, freshTenantID)
		assertRuntimePrimaryTenantTarget(
			t, ctx, database, "saas", "saas_control",
		)
	})
}

func runPlatformMigrations(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	runPlatformMigrationsThrough(t, ctx, database, origin, 0)
}

func runPlatformMigrationsThrough(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
	maxVersion int64,
) {
	t.Helper()
	connection := tenantMigratorConnection(t, ctx, database)
	defer connection.Close()
	directory := platformMigrationDirectory(origin)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), "."+origin+".sql") {
			t.Fatalf("unexpected %s migration: %s", origin, entry.Name())
		}
		version, err := strconv.ParseInt(entry.Name()[:14], 10, 64)
		if err != nil {
			t.Fatalf("parse migration version %s: %v", entry.Name(), err)
		}
		if maxVersion > 0 && version > maxVersion {
			continue
		}
		content, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := connection.ExecContext(ctx, string(content)); err != nil {
			t.Fatalf("apply %s: %v", entry.Name(), err)
		}
	}
}

func resetTenantMigrationCatalog(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	statements := []string{
		"DROP VIEW IF EXISTS public.tenants",
		"DROP TABLE IF EXISTS public.tenants",
		"DROP SCHEMA IF EXISTS platform CASCADE",
		"CREATE SCHEMA platform",
		`CREATE TABLE platform.environment_identity (
singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
environment_id uuid NOT NULL UNIQUE,
database_origin text NOT NULL CHECK (database_origin IN ('fresh','v1_adopted')),
database_system_identifier text NOT NULL,
database_oid bigint NOT NULL,
release_id uuid NOT NULL,
bootstrap_authority_sha256 text NOT NULL,
created_at timestamptz NOT NULL DEFAULT statement_timestamp())`,
		`INSERT INTO platform.environment_identity (
environment_id,database_origin,database_system_identifier,database_oid,
release_id,bootstrap_authority_sha256)
VALUES ('10000000-0000-4000-8000-000000000001','fresh','1',1,
'20000000-0000-4000-8000-000000000001',repeat('a',64))`,
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("reset tenant catalog: %v", err)
		}
	}
}

func insertFreshTenant(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	schemaName, err := domain.SchemaNameFromUUID(freshTenantID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `
INSERT INTO platform.tenants (
    id, name, slug, status, settings, owner_id, schema_name, created_at, updated_at
) VALUES ($1, 'Fresh', 'fresh', 'ACTIVE', '{}', $2, $3,
    TIMESTAMPTZ '2026-07-13 09:30:00+00',
    TIMESTAMPTZ '2026-07-13 09:30:00+00')`,
		freshTenantID, testOwnerID, schemaName.String())
	if err != nil {
		t.Fatal(err)
	}
}
