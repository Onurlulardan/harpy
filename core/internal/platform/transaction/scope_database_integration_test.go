//go:build integration

package transaction

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

const scopeIntegrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

const scopeRuntimePassword = "scope-runtime-integration-only"

type scopeTestDatabase struct {
	admin      *pgx.Conn
	runtimeDSN string
}

func startScopeTestDatabase(t *testing.T, ctx context.Context) scopeTestDatabase {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, scopeIntegrationImage,
		tcpostgres.WithDatabase("harpy_scope"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("scope-admin-integration-only"),
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
		t.Fatalf("start exact PostgreSQL 18.4 scope image: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate PostgreSQL scope container: %v", err)
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
	prepareScopeDatabase(t, ctx, admin)
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword("harpy_v2_runtime", scopeRuntimePassword)
	return scopeTestDatabase{admin: admin, runtimeDSN: parsed.String()}
}
