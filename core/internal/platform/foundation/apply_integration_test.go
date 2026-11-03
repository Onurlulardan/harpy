//go:build integration

package foundation

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestApplyIsIdempotentAndLeavesNoApplicationObjects(t *testing.T) {
	conn, ctx := adminConn(t)
	cfg := testConfig(databaseName(t))
	if err := Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	before := catalogSnapshot(t, ctx, conn, cfg.Database)
	if err := Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	after := catalogSnapshot(t, ctx, conn, cfg.Database)
	if before != after {
		t.Fatalf("catalog changed on second apply: %s -> %s", before, after)
	}
	if strings.Contains(before, "goose_db_version") || !strings.Contains(before, "relations=0") {
		t.Fatal(before)
	}
}

func TestApplyCompletesAfterOnlyTheOwnerRoleExists(t *testing.T) {
	conn, ctx := adminConn(t)
	cfg := testConfig(databaseName(t))
	if err := ensureRole(ctx, conn, cfg, managedRoles(cfg)[0]); err != nil {
		t.Fatal(err)
	}
	if err := Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if err := Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestApplyRefusesDirtyDatabaseWithoutDeleting(t *testing.T) {
	conn, ctx := adminConn(t)
	cfg := testConfig(databaseName(t))
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+ident(cfg.Database)); err != nil {
		t.Fatal(err)
	}
	user, password := adminIdentity(t)
	target := connectRole(t, ctx, user, password, cfg.Database)
	if _, err := target.Exec(ctx, "CREATE TABLE public.legacy(id integer)"); err != nil {
		t.Fatal(err)
	}
	if _, err := target.Exec(ctx, "INSERT INTO public.legacy(id) VALUES (7)"); err != nil {
		t.Fatal(err)
	}
	err := Apply(ctx, cfg)
	if err == nil || !strings.Contains(err.Error(), "public.legacy") {
		t.Fatal(err)
	}
	for _, secret := range []string{goosePassword, apiPassword, workerPassword} {
		if strings.Contains(err.Error(), secret) {
			t.Fatal("error contains a role password")
		}
	}
	var id int
	if err := target.QueryRow(ctx, "SELECT id FROM public.legacy").Scan(&id); err != nil || id != 7 {
		t.Fatal(err)
	}
	var platform bool
	if err := target.QueryRow(ctx, `SELECT EXISTS(
SELECT 1 FROM pg_namespace WHERE nspname = 'platform')`).Scan(&platform); err != nil || platform {
		t.Fatalf("platform schema created: %v %t", err, platform)
	}
	var owner bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS(
SELECT 1 FROM pg_roles WHERE rolname = $1)`, RoleOwner).Scan(&owner); err != nil || owner {
		t.Fatalf("owner role created: %v %t", err, owner)
	}
}

func catalogSnapshot(t *testing.T, ctx context.Context, conn *pgx.Conn, database string) string {
	t.Helper()
	targetDSN, err := databaseDSN(adminDSN, database)
	if err != nil {
		t.Fatal(err)
	}
	target, err := connect(ctx, targetDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close(ctx)
	var relations, goose int
	var extensions, owner string
	if err := target.QueryRow(ctx, `SELECT count(*) FROM pg_class relation
JOIN pg_namespace namespace ON namespace.oid = relation.relnamespace
WHERE namespace.nspname NOT IN ('pg_catalog', 'information_schema')
  AND namespace.nspname NOT LIKE 'pg\_%' ESCAPE '\'
  AND relation.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')`).Scan(&relations); err != nil {
		t.Fatal(err)
	}
	if err := target.QueryRow(ctx, `SELECT count(*) FROM pg_class
WHERE relname = 'goose_db_version'`).Scan(&goose); err != nil {
		t.Fatal(err)
	}
	if err := target.QueryRow(ctx, `SELECT coalesce(string_agg(extname, ',' ORDER BY extname), '')
FROM pg_extension`).Scan(&extensions); err != nil {
		t.Fatal(err)
	}
	if err := target.QueryRow(ctx, `SELECT pg_get_userbyid(nspowner) FROM pg_namespace
WHERE nspname = 'platform'`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if owner != RoleOwner || goose != 0 || !strings.Contains(extensions, "vector") {
		t.Fatalf("owner=%s goose=%d extensions=%s", owner, goose, extensions)
	}
	return "relations=" + itoa(relations) + " extensions=" + extensions
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	return "nonzero"
}
