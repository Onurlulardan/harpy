//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const supportPostgresImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

func startSupportDatabase(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	container, err := tcpostgres.Run(ctx, supportPostgresImage,
		tcpostgres.WithDatabase("harpy_support"),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("harpy_support_integration_only"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp")).WithStartupTimeout(90*time.Second)),
		testcontainers.WithEnv(map[string]string{"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 --data-checksums " +
			"--auth-local=trust --auth-host=scram-sha-256"}),
		testcontainers.WithCmd("postgres"))
	if err != nil {
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		t.Fatalf("start exact PostgreSQL 18.4 image: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate support PostgreSQL container: %v", err)
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
	if err = database.QueryRowContext(ctx, "SHOW server_version_num").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 180004 {
		t.Fatalf("PostgreSQL version = %d, want 180004", version)
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

func resetAndApplySupport(t *testing.T, ctx context.Context, database *sql.DB, origin string) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `DROP SCHEMA IF EXISTS platform CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx,
		`CREATE SCHEMA platform; GRANT USAGE ON SCHEMA platform TO harpy_runtime, harpy_adopter`); err != nil {
		t.Fatal(err)
	}
	documents, err := support.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	needle := "/migrations/platform/" + origin + "/"
	for _, document := range documents {
		if !strings.Contains(document.Path(), needle) {
			continue
		}
		if _, err = database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
	}
}
