//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertRuntimeTenantResolver(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID string,
) {
	t.Helper()
	var secure bool
	var settings, definition string
	err := database.QueryRowContext(ctx, `
SELECT routine.prosecdef,
       COALESCE(pg_catalog.array_to_string(routine.proconfig,','),''),
       pg_catalog.pg_get_functiondef(routine.oid)
FROM pg_catalog.pg_proc AS routine
WHERE routine.oid='platform.tenancy_runtime_resolve_command_v1(uuid)'::regprocedure`,
	).Scan(&secure, &settings, &definition)
	if err != nil || !secure || settings != "search_path=pg_catalog" ||
		!strings.Contains(definition, "FOR SHARE") {
		t.Fatalf("runtime resolver contract = %t %q: %v", secure, settings, err)
	}
	var expectedSchema, expectedSlug, expectedStatus string
	var expectedLive bool
	err = database.QueryRowContext(ctx, `SELECT schema_name,slug,status,deleted_at IS NULL
FROM platform.tenants WHERE id=$1`, tenantID).Scan(
		&expectedSchema, &expectedSlug, &expectedStatus, &expectedLive,
	)
	if err != nil {
		t.Fatal(err)
	}
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var schema, slug, status string
	var live bool
	err = transaction.QueryRowContext(ctx, `SELECT schema_name,slug,status,not_deleted
FROM platform.tenancy_runtime_resolve_command_v1($1)`, tenantID).Scan(
		&schema, &slug, &status, &live,
	)
	if err != nil || schema != expectedSchema || slug != expectedSlug ||
		status != expectedStatus || live != expectedLive {
		t.Fatalf("runtime resolver = %q %q %q %t: %v", schema, slug, status, live, err)
	}
	if _, err := transaction.ExecContext(ctx,
		"UPDATE platform.tenants SET id=id WHERE id=$1", tenantID,
	); err == nil {
		t.Fatal("runtime role unexpectedly updates the tenant registry")
	}
}
