//go:build integration

package goose

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	"github.com/jackc/pgx/v5"
)

func assertGooseHistoryCount(t *testing.T, ctx context.Context, cfg foundation.Config, want int) {
	t.Helper()
	var count int
	err := gooseAdmin(t, cfg).QueryRow(ctx, `SELECT count(*)
FROM pg_catalog.pg_class WHERE relname = 'goose_db_version'`).Scan(&count)
	if err != nil || count != want {
		t.Fatalf("goose history count=%d want=%d err=%v", count, want, err)
	}
}

func assertTenantHistories(
	t *testing.T, ctx context.Context, cfg foundation.Config, first, second [16]byte,
) {
	t.Helper()
	left, err := foundation.TenantSchemaName(first)
	if err != nil {
		t.Fatal(err)
	}
	right, err := foundation.TenantSchemaName(second)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := gooseAdmin(t, cfg).Query(ctx, `SELECT namespace.nspname
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE relation.relname = 'goose_db_version' ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	want := []string{foundation.PlatformSchema, left, right}
	if len(names) != len(want) {
		t.Fatalf("history schemas=%v", names)
	}
	seen := map[string]bool{}
	for _, name := range names {
		seen[name] = true
	}
	for _, name := range want {
		if !seen[name] {
			t.Fatalf("history schemas=%v missing %s", names, name)
		}
	}
}

func assertRuntimeDDLDenied(t *testing.T, ctx context.Context, cfg foundation.Config) {
	t.Helper()
	for _, role := range []struct{ name, password string }{
		{foundation.RoleAPI, cfg.APIPassword},
		{foundation.RoleWorker, cfg.WorkerPassword},
	} {
		dsn, err := userDSN(cfg.AdminDSN, role.name, role.password, cfg.Database)
		if err != nil {
			t.Fatal(err)
		}
		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			t.Fatal(err)
		}
		_, err = conn.Exec(ctx, "CREATE TABLE platform.runtime_must_not_create(id integer)")
		_ = conn.Close(ctx)
		if err == nil {
			t.Fatalf("%s created a table", role.name)
		}
	}
}

func setRolePassword(
	t *testing.T, ctx context.Context, cfg foundation.Config, role, password string,
) error {
	t.Helper()
	var statement string
	err := gooseAdmin(t, cfg).QueryRow(ctx,
		"SELECT pg_catalog.format('ALTER ROLE %I PASSWORD %L', $1::text, $2::text)",
		role, password,
	).Scan(&statement)
	if err != nil {
		return err
	}
	_, err = gooseAdmin(t, cfg).Exec(ctx, statement)
	return err
}
