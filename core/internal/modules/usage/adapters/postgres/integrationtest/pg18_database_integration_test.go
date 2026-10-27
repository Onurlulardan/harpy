//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage"
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
		tcpostgres.WithDatabase("harpy_usage"),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("harpy_usage_integration_only"),
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
			t.Errorf("terminate usage PostgreSQL container: %v", err)
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
	if _, err = database.ExecContext(ctx,
		`CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER`); err != nil {
		t.Fatal(err)
	}
	if _, err = database.ExecContext(ctx,
		`CREATE ROLE harpy_adopter NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0`); err != nil {
		t.Fatal(err)
	}
	return database
}

func resetAndApplyUsage(t *testing.T, ctx context.Context, database *sql.DB, origin string) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `DROP SCHEMA IF EXISTS platform CASCADE;
CREATE SCHEMA platform; GRANT USAGE ON SCHEMA platform TO harpy_runtime,harpy_adopter`); err != nil {
		t.Fatal(err)
	}
	documents, err := usage.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	needle := "/migrations/platform/" + origin + "/"
	for _, document := range documents {
		if strings.Contains(document.Path(), needle) {
			if _, err = database.ExecContext(ctx, string(document.Content())); err != nil {
				t.Fatalf("apply %s: %v", document.Path(), err)
			}
		}
	}
}

func TestUsageMigrationsApplyOnPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startUsageDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) { resetAndApplyUsage(t, ctx, database, origin) })
	}
}
