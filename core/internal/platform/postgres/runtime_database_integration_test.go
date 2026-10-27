//go:build integration

package postgres

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const runtimeIntegrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

const runtimeIntegrationPassword = "runtime-integration-only"

type runtimeTestDatabase struct {
	admin       *pgx.Conn
	runtimeDSN  string
	migratorDSN string
}

func startRuntimeTestDatabase(t *testing.T, ctx context.Context) runtimeTestDatabase {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, runtimeIntegrationImage,
		tcpostgres.WithDatabase("harpy_runtime"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("admin-integration-only"),
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
	adminDSN, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })
	prepareRuntimeRoles(t, ctx, admin)
	prepareRuntimeCatalog(t, ctx, admin)
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword("harpy_v2_runtime", runtimeIntegrationPassword)
	runtimeDSN := parsed.String()
	parsed.User = url.UserPassword("harpy_migrator", runtimeIntegrationPassword)
	return runtimeTestDatabase{
		admin: admin, runtimeDSN: runtimeDSN, migratorDSN: parsed.String(),
	}
}
