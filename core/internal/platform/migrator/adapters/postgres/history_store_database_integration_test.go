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

const (
	historyIntegrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
		"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"
	historyIntegrationDatabase = "harpy_history"
	historyIntegrationUser     = "harpy_migrator"
	historyIntegrationPassword = "harpy_history_integration_only"
)

func startHistoryIntegrationDatabase(
	t *testing.T,
	ctx context.Context,
) *sql.DB {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx,
		historyIntegrationImage,
		tcpostgres.WithDatabase(historyIntegrationDatabase),
		tcpostgres.WithUsername(historyIntegrationUser),
		tcpostgres.WithPassword(historyIntegrationPassword),
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
		t.Fatalf("start exact PostgreSQL 18.4 integration image: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate PostgreSQL integration container: %v", err)
		}
	})
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("resolve PostgreSQL integration connection: %v", err)
	}
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open PostgreSQL integration connection: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close PostgreSQL integration connection: %v", err)
		}
	})
	if err := database.PingContext(ctx); err != nil {
		t.Fatalf("ping PostgreSQL integration database: %v", err)
	}
	assertPostgreSQL184(t, ctx, database)
	prepareHistoryIntegrationRoles(t, ctx, database)
	return database
}

func assertPostgreSQL184(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var versionNumber int
	if err := database.QueryRowContext(ctx, "SHOW server_version_num").
		Scan(&versionNumber); err != nil {
		t.Fatalf("read PostgreSQL server version: %v", err)
	}
	if versionNumber != 180004 {
		t.Fatalf("PostgreSQL server version = %d, want 180004", versionNumber)
	}
}

func prepareHistoryIntegrationRoles(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	statements := []string{
		"CREATE ROLE harpy_owner NOLOGIN",
		"GRANT harpy_owner TO harpy_migrator",
		"ALTER DATABASE harpy_history OWNER TO harpy_owner",
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("prepare canonical migration roles: %v", err)
		}
	}
}
