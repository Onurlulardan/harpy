//go:build integration

package postgres

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const cutoverIntegrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

type cutoverIntegrationDatabase struct {
	dsn   string
	admin *pgx.Conn
}

func startCutoverIntegrationDatabase(
	t *testing.T, ctx context.Context) cutoverIntegrationDatabase {
	t.Helper()
	return startCutoverIntegrationDatabaseForOrigin(t, ctx, "fresh")
}

func startCutoverIntegrationDatabaseForOrigin(
	t *testing.T, ctx context.Context, origin string,
) cutoverIntegrationDatabase {
	return startCutoverIntegrationDatabaseForOriginSkipping(t, ctx, origin, nil)
}

func startCutoverIntegrationDatabaseForOriginSkipping(t *testing.T, ctx context.Context,
	origin string, skipped map[string]bool) cutoverIntegrationDatabase {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("unsupported cutover migration origin %q", origin)
	}
	container, err := tcpostgres.Run(
		ctx, cutoverIntegrationImage,
		tcpostgres.WithDatabase("harpy_cutover"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("cutover-integration-only"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "scram-sha-256",
			"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 " +
				"--data-checksums --auth-local=scram-sha-256 --auth-host=scram-sha-256",
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
		t.Fatalf("start exact PostgreSQL 18.4 cutover database: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate cutover database: %v", err)
		}
	})
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })
	var version int
	if err := admin.QueryRow(ctx,
		"SELECT current_setting('server_version_num')::integer").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 180004 {
		t.Fatalf("PostgreSQL version = %d, want 180004", version)
	}
	prepareCutoverCatalog(t, ctx, admin)
	migrator := openCutoverMigrator(t, ctx, dsn)
	if _, err := migrator.Exec(ctx, "SET ROLE harpy_owner; CREATE SCHEMA platform"); err != nil {
		t.Fatal(err)
	}
	prepareReadinessSurfaces(t, ctx, migrator, origin)
	applyCutoverMigrations(t, ctx, migrator, origin, skipped)
	assertCutoverFunctionAuthority(t, ctx, admin)
	if err := migrator.Close(ctx); err != nil {
		t.Fatal(err)
	}
	return cutoverIntegrationDatabase{dsn: dsn, admin: admin}
}

func prepareCutoverCatalog(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	_, err := admin.Exec(ctx, `
CREATE ROLE harpy_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0 PASSWORD NULL;
CREATE ROLE harpy_cutover_authority_owner NOLOGIN NOINHERIT NOSUPERUSER
NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT -1
PASSWORD NULL VALID UNTIL 'infinity';
ALTER DATABASE harpy_cutover OWNER TO harpy_owner;
CREATE ROLE harpy_migrator LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT -1 PASSWORD 'integration-migrator-secret';
CREATE ROLE harpy_cluster_admin NOLOGIN NOINHERIT NOSUPERUSER CREATEDB CREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0 PASSWORD NULL;
GRANT harpy_owner TO harpy_migrator WITH INHERIT FALSE, SET TRUE;
CREATE ROLE harpy_adopter NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL;
CREATE ROLE harpy_cutover_decoder NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL;
CREATE ROLE harpy_cutover_finalizer NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL;
CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0 PASSWORD NULL;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT -1 PASSWORD 'integration-runtime-secret';
GRANT harpy_runtime TO harpy_v2_runtime WITH INHERIT FALSE, SET TRUE;
REVOKE TEMPORARY ON DATABASE harpy_cutover FROM PUBLIC;`)
	if err != nil {
		t.Fatalf("prepare cutover role catalog: %v", err)
	}
}

func openCutoverMigrator(t *testing.T, ctx context.Context, dsn string) *pgx.Conn {
	t.Helper()
	configuration, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	configuration.User = "harpy_migrator"
	configuration.Password = "integration-migrator-secret"
	connection, err := pgx.ConnectConfig(ctx, configuration)
	if err != nil {
		t.Fatal(err)
	}
	return connection
}

func applyCutoverMigrations(
	t *testing.T, ctx context.Context, admin *pgx.Conn, origin string,
	skipped map[string]bool,
) {
	t.Helper()
	documents, err := platformcutover.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(documents, func(left, right int) bool {
		return filepath.Base(documents[left].Path()) <
			filepath.Base(documents[right].Path())
	})
	for _, document := range documents {
		name := filepath.Base(document.Path())
		if !strings.HasSuffix(name, "."+origin+".sql") || skipped[name] {
			continue
		}
		applyCutoverMigrationPrerequisites(t, ctx, admin, origin, name)
		if _, err := admin.Exec(ctx, string(document.Content())); err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) {
				t.Fatalf("apply cutover migration %s: SQLSTATE=%s position=%d: %v",
					name, postgresError.Code, postgresError.Position, err)
			}
			t.Fatalf("apply cutover migration %s: %v", name, err)
		}
	}
}

func connectCutoverRole(
	t *testing.T,
	ctx context.Context,
	dsn string,
	role string,
) *pgx.Conn {
	t.Helper()
	configuration, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	configuration.User = role
	configuration.Password = "integration-leased-secret"
	configuration.RuntimeParams["search_path"] = "pg_catalog"
	connection, err := pgx.ConnectConfig(ctx, configuration)
	if err != nil {
		t.Fatalf("connect %s: %v", role, err)
	}
	return connection
}
