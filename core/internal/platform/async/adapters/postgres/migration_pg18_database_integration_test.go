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

const asyncMigrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

func startAsyncMigrationDatabase(
	t *testing.T,
	ctx context.Context,
	origin string,
) *sql.DB {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx,
		asyncMigrationImage,
		tcpostgres.WithDatabase("harpy_async_"+origin),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("harpy_async_integration_only"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(1),
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
		t.Fatalf("start exact PostgreSQL 18.4 async database: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate async PostgreSQL container: %v", err)
		}
	})
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("resolve async PostgreSQL connection: %v", err)
	}
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open async PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close async PostgreSQL connection: %v", err)
		}
	})
	var version int
	if err := database.QueryRowContext(ctx, "SHOW server_version_num").
		Scan(&version); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	if version != 180004 {
		t.Fatalf("PostgreSQL version = %d, want 180004", version)
	}
	return database
}
