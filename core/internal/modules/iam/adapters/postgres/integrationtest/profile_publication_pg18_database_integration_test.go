//go:build integration

package integrationtest

import (
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const publicationPGImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

const publicationPassword = "profile-publication-integration-only"

type publicationDatabase struct {
	admin      *pgx.Conn
	runtimeDSN string
	tenant     publicationTenant
}

type publicationTenant struct {
	id, schema, origin string
}

func startPublicationPG18(
	t *testing.T,
	ctx context.Context,
	origin string,
) publicationDatabase {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, publicationPGImage,
		tcpostgres.WithDatabase("harpy_publication_"+origin),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("publication-admin-only"),
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
		t.Fatalf("start PostgreSQL 18.4 publication database: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })
	var versionText string
	if err := admin.QueryRow(ctx, "SHOW server_version_num").Scan(&versionText); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	version, err := strconv.Atoi(versionText)
	if err != nil || version != 180004 {
		t.Fatalf("PostgreSQL version = %q/%v, want 180004", versionText, err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword("harpy_v2_runtime", publicationPassword)
	return publicationDatabase{
		admin: admin, runtimeDSN: parsed.String(),
		tenant: publicationTenant{
			id:     "10000000-0000-4000-8000-000000000001",
			schema: "t_profile_publication", origin: origin,
		},
	}
}
