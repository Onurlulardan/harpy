//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const tenantMigrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

func startTenantMigrationDatabase(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx,
		tenantMigrationImage,
		tcpostgres.WithDatabase("harpy_tenancy"),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("harpy_tenancy_integration_only"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(90*time.Second)),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 " +
				"--data-checksums --auth-local=trust --auth-host=scram-sha-256",
		}),
		testcontainers.WithCmd("postgres"),
	)
	if err != nil {
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		t.Fatalf("start exact PostgreSQL 18.4 image: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate PostgreSQL container: %v", err)
		}
	})
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	var version int
	if err := database.QueryRowContext(ctx, "SHOW server_version_num").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 180004 {
		t.Fatalf("PostgreSQL version = %d, want 180004", version)
	}
	if _, err := database.ExecContext(ctx, `CREATE ROLE harpy_runtime
NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
		t.Fatalf("create runtime role: %v", err)
	}
	if _, err := database.ExecContext(ctx, `CREATE ROLE harpy_v2_runtime
LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
		t.Fatalf("create V2 runtime role: %v", err)
	}
	if _, err := database.ExecContext(ctx, `CREATE ROLE harpy_tenant_migration_executor
NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
		t.Fatalf("create tenant migration executor role: %v", err)
	}
	if _, err := database.ExecContext(ctx, `CREATE ROLE harpy_backup_reader
NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
		t.Fatalf("create backup reader role: %v", err)
	}
	if _, err := database.ExecContext(ctx, `CREATE ROLE harpy_restore
LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
		t.Fatalf("create restore role: %v", err)
	}
	if _, err := database.ExecContext(ctx, `CREATE ROLE harpy_adopter
NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
		t.Fatalf("create adopter role: %v", err)
	}
	if _, err := database.ExecContext(ctx, `CREATE ROLE harpy_migrator
LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
		t.Fatalf("create migrator role: %v", err)
	}
	if _, err := database.ExecContext(ctx, `GRANT harpy_owner TO harpy_migrator`); err != nil {
		t.Fatalf("grant owner role to migrator: %v", err)
	}
	if _, err := database.ExecContext(
		ctx, "GRANT harpy_runtime TO harpy_v2_runtime",
	); err != nil {
		t.Fatalf("grant runtime role: %v", err)
	}
	return database
}
