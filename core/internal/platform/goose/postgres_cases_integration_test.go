//go:build integration

package goose

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	"github.com/jackc/pgx/v5"
)

func TestPlatformGooseCreatesTenantsAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	cfg := gooseConfig("goose_platform_case")
	if err := foundation.Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if err := UpPlatform(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if err := UpPlatform(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	conn := gooseAdmin(t, cfg)
	var relations int
	err := conn.QueryRow(ctx, `SELECT count(*) FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = 'platform' AND relation.relname = 'tenants'`).Scan(&relations)
	if err != nil || relations != 1 {
		t.Fatalf("tenants relations=%d err=%v", relations, err)
	}
	var historySchema string
	err = conn.QueryRow(ctx, `SELECT namespace.nspname
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE relation.relname = 'goose_db_version'`).Scan(&historySchema)
	if err != nil || historySchema != "platform" {
		t.Fatalf("history schema=%s err=%v", historySchema, err)
	}
}

func TestUpPreservesPostgresError(t *testing.T) {
	ctx := context.Background()
	cfg := gooseConfig("goose_error_case")
	if err := foundation.Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	dsn, err := gooseDSN(cfg)
	if err != nil {
		t.Fatal(err)
	}
	source := fstest.MapFS{"00099_bad.sql": {Data: []byte(
		"-- +goose Up\nSELECT missing_goose_relation;\n-- +goose Down\nSELECT 1;\n",
	)}}
	err = up(ctx, dsn, source, foundation.RoleOwner, foundation.PlatformSchema)
	if err == nil || !strings.Contains(err.Error(), "missing_goose_relation") {
		t.Fatal(err)
	}
}

func TestTenantGooseIsolatesHistory(t *testing.T) {
	ctx := context.Background()
	cfg := gooseConfig("goose_tenant_case")
	if err := foundation.Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if err := UpPlatform(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	first := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	second := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	if err := ProvisionTenant(ctx, cfg, first); err != nil {
		t.Fatal(err)
	}
	if err := ProvisionTenant(ctx, cfg, first); err != nil {
		t.Fatal(err)
	}
	if err := ProvisionTenant(ctx, cfg, second); err != nil {
		t.Fatal(err)
	}
	conn := gooseAdmin(t, cfg)
	var schemas int
	err := conn.QueryRow(ctx, `SELECT count(DISTINCT namespace.nspname)
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE relation.relname = 'goose_db_version' AND namespace.nspname <> 'platform'`).Scan(&schemas)
	if err != nil || schemas != 2 {
		t.Fatalf("tenant histories=%d err=%v", schemas, err)
	}
	apiDSN, err := userDSN(cfg.AdminDSN, foundation.RoleAPI, cfg.APIPassword, cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	api, err := pgx.Connect(ctx, apiDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer api.Close(ctx)
	_, err = api.Exec(ctx, "CREATE TABLE platform.api_must_not_create(id integer)")
	if err == nil {
		t.Fatal("api role created a table")
	}
}

func foundationTargetDSN(cfg foundation.Config) (string, error) {
	parsed, err := url.Parse(cfg.AdminDSN)
	if err != nil {
		return "", err
	}
	parsed.Path = "/" + cfg.Database
	return parsed.String(), nil
}

func gooseAdmin(t *testing.T, cfg foundation.Config) *pgx.Conn {
	t.Helper()
	dsn, err := foundationTargetDSN(cfg)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}
