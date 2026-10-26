//go:build integration

package integrationtest

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

const catalogPG18Image = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

type catalogPG18 struct {
	owner *sql.DB
	dsn   string
}

func startCatalogPG18(t *testing.T, ctx context.Context, origin string) catalogPG18 {
	t.Helper()
	container, err := tcpostgres.Run(ctx, catalogPG18Image,
		tcpostgres.WithDatabase("harpy_catalog_"+origin),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("catalog-integration-owner-only"),
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
		t.Fatalf("start PostgreSQL 18.4 Catalog database: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = owner.Close() })
	var version int
	if err = owner.QueryRowContext(ctx, "SHOW server_version_num").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 180004 {
		t.Fatalf("PostgreSQL version=%d, want 180004", version)
	}
	return catalogPG18{owner: owner, dsn: dsn}
}
