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

const layoutPG18Image = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

func startLayoutPG18(t *testing.T, ctx context.Context, origin string) *sql.DB {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, layoutPG18Image,
		tcpostgres.WithDatabase("harpy_layout_"+origin),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("layout_integration_only"),
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
		t.Fatalf("start PostgreSQL 18.4 layout database: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate layout PostgreSQL: %v", err)
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
	return database
}
