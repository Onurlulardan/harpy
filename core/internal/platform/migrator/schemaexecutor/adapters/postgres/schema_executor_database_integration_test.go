//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const schemaExecutorIntegrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

const schemaExecutorMigratorPassword = "schema-executor-migrator-only"

type schemaExecutorIntegrationEnvironment struct {
	admin *pgxpool.Pool
	pool  *pgxpool.Pool
	store *schemaExecutorStore
	dsn   string
}

func startSchemaExecutorIntegration(
	t *testing.T,
	ctx context.Context,
) *schemaExecutorIntegrationEnvironment {
	return startSchemaExecutorIntegrationVariant(t, ctx, "fresh")
}

func startSchemaExecutorIntegrationVariant(
	t *testing.T,
	ctx context.Context,
	variant string,
) *schemaExecutorIntegrationEnvironment {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, schemaExecutorIntegrationImage,
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername("entrypoint_test_admin"),
		tcpostgres.WithPassword("adapter_container_only"),
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
		t.Fatalf("start exact PostgreSQL 18.4 schema executor database: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate schema executor database: %v", err)
		}
	})
	prepareSchemaExecutorRoles(t, ctx, container)
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	adminConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse schema executor fixture admin endpoint: %v", err)
	}
	adminConfig.ConnConfig.Database = "harpy_adapter"
	admin, err := pgxpool.NewWithConfig(ctx, adminConfig)
	if err != nil {
		t.Fatalf("open schema executor fixture admin: %v", err)
	}
	t.Cleanup(admin.Close)
	database, pool := openSchemaExecutorDatabases(t, ctx, dsn)
	assertSchemaExecutorPostgreSQL184(t, ctx, pool)
	applySchemaExecutorMigrationsVariant(t, ctx, database, variant)
	environment := &schemaExecutorIntegrationEnvironment{
		admin: admin, pool: pool,
		store: &schemaExecutorStore{pool: pool, cleanupTimeout: 5 * time.Second},
		dsn:   dsn,
	}
	prepareSchemaExecutorTrigramRegistry(t, ctx, environment, variant)
	return environment
}

func assertSchemaExecutorPostgreSQL184(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()
	var version, versionNumber string
	if err := pool.QueryRow(ctx, "SHOW server_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SHOW server_version_num").Scan(&versionNumber); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(version, "18.4") {
		t.Fatalf("PostgreSQL version = %s, want 18.4", version)
	}
	if versionNumber != "180004" {
		t.Fatalf("PostgreSQL version number = %s, want 180004", versionNumber)
	}
}

func openSchemaExecutorDatabases(
	t *testing.T,
	ctx context.Context,
	dsn string,
) (*sql.DB, *pgxpool.Pool) {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.Database, config.User, config.Password = "harpy_adapter",
		"harpy_migrator", schemaExecutorMigratorPassword
	config.RuntimeParams["role"] = "harpy_owner"
	config.RuntimeParams["search_path"] = "public,pg_catalog"
	database := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = database.Close() })
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.ConnConfig.Database = "harpy_adapter"
	poolConfig.ConnConfig.User = "harpy_migrator"
	poolConfig.ConnConfig.Password = schemaExecutorMigratorPassword
	poolConfig.ConnConfig.RuntimeParams = map[string]string{
		"role": "harpy_owner", "search_path": "pg_catalog",
	}
	poolConfig.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return database, pool
}
