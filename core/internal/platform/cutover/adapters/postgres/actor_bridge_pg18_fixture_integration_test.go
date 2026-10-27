//go:build integration

package postgres

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	actorBridgeTenantA = "31000000-0000-4000-8000-000000000001"
	actorBridgeTenantB = "32000000-0000-4000-8000-000000000002"
	actorBridgeSchemaA = "t_actor_a"
	actorBridgeSchemaB = "t_actor_b"
)

type actorBridgeTenantFixture struct {
	id, schema string
}

var actorBridgeTenants = []actorBridgeTenantFixture{
	{id: actorBridgeTenantA, schema: actorBridgeSchemaA},
	{id: actorBridgeTenantB, schema: actorBridgeSchemaB},
}

func prepareActorBridgeIAM(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase, origin string,
) {
	t.Helper()
	prepareActorBridgeRolesAndSchemas(t, ctx, database.admin, origin)
	migrator := openCutoverMigrator(t, ctx, database.dsn)
	defer migrator.Close(ctx)
	if _, err := migrator.Exec(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	for _, tenant := range actorBridgeTenants {
		applyActorBridgeIAMDocuments(t, ctx, migrator, origin, tenant.schema)
		seedActorBridgeIAMTenant(t, ctx, migrator, tenant)
	}
	seedActorBridgeAuthority(t, ctx, database.admin)
	if _, err := migrator.Exec(ctx,
		"SELECT * FROM platform.iam_v1_actor_bridge_reconcile_v1(1)",
	); err != nil {
		t.Fatalf("reconcile actor bridge retirement: %v", err)
	}
}

func prepareActorBridgeRolesAndSchemas(
	t *testing.T, ctx context.Context, admin *pgx.Conn, origin string,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `
CREATE ROLE harpy_v1_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT -1
PASSWORD 'integration-v1-secret';
GRANT harpy_runtime TO harpy_v1_runtime WITH INHERIT FALSE, SET TRUE;
CREATE SCHEMA t_actor_a AUTHORIZATION harpy_owner;
CREATE SCHEMA t_actor_b AUTHORIZATION harpy_owner;
	INSERT INTO platform.tenants (id, schema_name, deleted_at) VALUES
	    ($1,$2,NULL),($3,$4,NULL);`,
		pgx.QueryExecModeSimpleProtocol,
		actorBridgeTenantA, actorBridgeSchemaA, actorBridgeTenantB, actorBridgeSchemaB)
	if err != nil {
		t.Fatalf("prepare actor bridge roles and schemas: %v", err)
	}
	if origin == "v1_adopted" {
		seedActorBridgeLegacySentinels(t, ctx, admin)
	}
}

func applyActorBridgeIAMDocuments(
	t *testing.T, ctx context.Context, migrator *pgx.Conn, origin, schema string,
) {
	t.Helper()
	directory := filepath.Join("..", "..", "..", "..", "modules", "iam",
		"migrations", "tenant", origin)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, 0, 46)
	for _, entry := range entries {
		if entry.Type().IsRegular() && actorBridgeIAMMigrationSelected(entry.Name()) {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(paths)
	if len(paths) != 45 {
		t.Fatalf("actor bridge IAM migration count = %d, want 45", len(paths))
	}
	if _, err := migrator.Exec(ctx, "SET search_path = "+pgx.Identifier{schema}.Sanitize()+", pg_catalog"); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		document, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := migrator.Exec(ctx, string(document)); err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) {
				t.Fatalf("apply %s: SQLSTATE=%s position=%d: %v",
					filepath.Base(path), postgresError.Code, postgresError.Position, err)
			}
			t.Fatalf("apply %s: %v", filepath.Base(path), err)
		}
	}
}

func actorBridgeIAMMigrationSelected(name string) bool {
	if !strings.HasSuffix(name, ".sql") || len(name) < 14 {
		return false
	}
	version, err := strconv.ParseInt(name[:14], 10, 64)
	if err != nil {
		return false
	}
	return version >= 20260716011000 && version <= 20260716012900 ||
		version >= 20260716012920 && version <= 20260716012926 ||
		version >= 20260716013020 && version <= 20260716013150 ||
		version == 20260716013433 ||
		version >= 20260716015800 && version <= 20260716015803
}

func connectActorBridgeV1(
	t *testing.T, ctx context.Context, dsn, schema string,
) *pgx.Conn {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword("harpy_v1_runtime", "integration-v1-secret")
	configuration, err := pgx.ParseConfig(parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	configuration.RuntimeParams["role"] = "harpy_runtime"
	configuration.RuntimeParams["search_path"] =
		pgx.Identifier{schema}.Sanitize() + ",pg_catalog"
	connection, err := pgx.ConnectConfig(ctx, configuration)
	if err != nil {
		t.Fatal(err)
	}
	return connection
}
