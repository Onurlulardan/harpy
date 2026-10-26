//go:build integration

package integrationfixture

import (
	"context"
	"database/sql"
	"io"
	"net/url"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const CatalogDemoTenantID = "76000000-0000-4000-8000-000000000071"
const CatalogDemoTenantSchema = "catalog_a"
const CatalogDemoOwnerGeneration int64 = 7
const CatalogDemoExtensionDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const catalogDemoImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

type CatalogDemoEnvironment struct {
	owner   *sql.DB
	runtime *sql.DB
	dsn     string
}

func StartCatalogDemoEnvironment(
	t *testing.T,
	ctx context.Context,
	root string,
	origin string,
	catalogDocuments []migration.Document,
	formsDocuments []migration.Document,
	asyncDocuments []migration.Document,
	analyticsStreams ...[]migration.Document,
) *CatalogDemoEnvironment {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("invalid Catalog demo origin %q", origin)
	}
	if len(analyticsStreams) > 1 {
		t.Fatal("Catalog demo accepts at most one Analytics migration stream")
	}
	var analyticsDocuments []migration.Document
	if len(analyticsStreams) == 1 {
		analyticsDocuments = analyticsStreams[0]
	}
	container, err := tcpostgres.Run(ctx, catalogDemoImage,
		tcpostgres.WithDatabase("harpy_catalog_demo_"+origin),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("catalog-demo-owner-only"),
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
		t.Fatalf("start PostgreSQL 18.4 Catalog demo database: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		logs, logErr := container.Logs(context.Background())
		if logErr != nil {
			t.Logf("read Catalog demo PostgreSQL logs: %v", logErr)
			return
		}
		defer logs.Close()
		content, readErr := io.ReadAll(logs)
		t.Logf("Catalog demo PostgreSQL logs (read error=%v):\n%s", readErr, content)
	})
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = owner.Close() })
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword("harpy_v2_runtime", "runtime-only")
	runtime, err := sql.Open("pgx", parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	environment := &CatalogDemoEnvironment{
		owner: owner, runtime: runtime, dsn: parsed.String(),
	}
	environment.prepare(t, ctx, root, origin,
		catalogDocuments, formsDocuments, analyticsDocuments, asyncDocuments)
	return environment
}

func (environment *CatalogDemoEnvironment) DSN() string {
	return environment.dsn
}
