//go:build integration

package adoption

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const usagePostgresImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

func startUsageDatabase(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	container, err := tcpostgres.Run(ctx, usagePostgresImage,
		tcpostgres.WithDatabase("harpy_usage_adoption"),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("harpy_usage_adoption_integration_only"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp")).WithStartupTimeout(90*time.Second)),
		testcontainers.WithEnv(map[string]string{"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 --data-checksums --auth-local=trust --auth-host=scram-sha-256"}),
		testcontainers.WithCmd("postgres"))
	if err != nil {
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		t.Fatalf("start exact PostgreSQL 18.4 image: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate usage adoption PostgreSQL container: %v", err)
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
	if err = database.QueryRowContext(ctx, "SHOW server_version_num").Scan(&version); err != nil ||
		version != 180004 {
		t.Fatalf("PostgreSQL version=%d err=%v, want 180004", version, err)
	}
	for _, statement := range []string{
		"CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER",
		"CREATE ROLE harpy_adopter NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0",
	} {
		if _, err = database.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	return database
}

func resetAndApplyUsage(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `DROP SCHEMA IF EXISTS platform CASCADE;
CREATE SCHEMA platform; GRANT USAGE ON SCHEMA platform TO harpy_runtime,harpy_adopter`); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join("..", "..", "..", "migrations", "platform", "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("unexpected usage migration directory %s", entry.Name())
		}
		document, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, err = database.ExecContext(ctx, string(document)); err != nil {
			t.Fatalf("apply %s: %v", entry.Name(), err)
		}
	}
}
