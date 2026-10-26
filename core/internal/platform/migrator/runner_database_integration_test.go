//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const runnerIntegrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

const runnerMigratorPassword = "runner-migrator-only"

func startRunnerIntegrationDatabase(
	t *testing.T,
	ctx context.Context,
) (*sql.DB, *pgx.ConnConfig) {
	return startRunnerIntegrationDatabaseWithConfig(t, ctx)
}

func startRunnerIntegrationDatabaseWithConfig(
	t *testing.T,
	ctx context.Context,
) (*sql.DB, *pgx.ConnConfig) {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx,
		runnerIntegrationImage,
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername("entrypoint_test_admin"),
		tcpostgres.WithPassword("runner_container_only"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "scram-sha-256",
			"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 " +
				"--data-checksums --auth-local=trust --auth-host=scram-sha-256",
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(90*time.Second)),
		testcontainers.WithCmd("postgres"),
	)
	if err != nil {
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		t.Fatalf("start exact PostgreSQL 18.4 runner database: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate runner database: %v", err)
		}
	})
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("resolve runner database endpoint: %v", err)
	}
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect runner fixture admin: %v", err)
	}
	prepareRunnerRoles(t, ctx, admin)
	assertProductionRoleFixture(t, ctx, admin)
	if _, err := admin.Exec(ctx,
		"ALTER ROLE harpy_migrator PASSWORD 'runner-migrator-only'"); err != nil {
		t.Fatalf("lease runner migrator test role: %v", err)
	}
	_ = admin.Close(context.Background())
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse runner database endpoint: %v", err)
	}
	config.Database = "harpy_runner"
	config.User = "harpy_migrator"
	config.Password = runnerMigratorPassword
	config.RuntimeParams["role"] = "harpy_owner"
	config.RuntimeParams["search_path"] = "pg_catalog"
	database := stdlib.OpenDB(*config)
	database.SetMaxOpenConns(4)
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close runner database: %v", err)
		}
	})
	if err := database.PingContext(ctx); err != nil {
		t.Fatalf("ping runner database: %v", err)
	}
	assertRunnerPostgreSQL184(t, ctx, database)
	return database, config.Copy()
}

func assertRunnerPostgreSQL184(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var version, versionNumber string
	if err := database.QueryRowContext(ctx, "SHOW server_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, "SHOW server_version_num").Scan(&versionNumber); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(version, "18.4") {
		t.Fatalf("PostgreSQL version = %s, want 18.4", version)
	}
	if versionNumber != "180004" {
		t.Fatalf("PostgreSQL version number = %s, want 180004", versionNumber)
	}
}

func prepareRunnerRoles(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
) {
	t.Helper()
	statements := productionRoleFixtureStatements("harpy_runner")
	statements = append(statements, "GRANT pg_signal_backend TO harpy_owner")
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("prepare runner production role fixture: %v", err)
		}
	}
}
