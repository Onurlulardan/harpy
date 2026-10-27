//go:build integration

package executor

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

const pg18Image = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

const pg18RuntimePassword = "restquery-runtime-integration-only"

type pg18Database struct {
	admin      *pgx.Conn
	runtimeDSN string
}

func pg18StartDatabase(t *testing.T, ctx context.Context) pg18Database {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, pg18Image,
		tcpostgres.WithDatabase("harpy_restquery"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("restquery-admin-integration-only"),
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
			t.Errorf("terminate PostgreSQL 18.4 container: %v", err)
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
	pg18PrepareSchema(t, ctx, admin)
	pg18PrepareFoundation(t, ctx, admin)
	pg18SeedTenants(t, ctx, admin)
	pg18AssertBaseline(t, ctx, admin)
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword("harpy_v2_runtime", pg18RuntimePassword)
	return pg18Database{admin: admin, runtimeDSN: parsed.String()}
}

func pg18AssertBaseline(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	var versionNumber, vectorVersion string
	if err := admin.QueryRow(ctx, `SELECT current_setting('server_version_num'),
(SELECT extversion FROM pg_catalog.pg_extension WHERE extname='vector')`).Scan(
		&versionNumber, &vectorVersion,
	); err != nil {
		t.Fatal(err)
	}
	if versionNumber != "180004" || vectorVersion != "0.8.3" {
		t.Fatalf("baseline PostgreSQL=%s vector=%s", versionNumber, vectorVersion)
	}
}
