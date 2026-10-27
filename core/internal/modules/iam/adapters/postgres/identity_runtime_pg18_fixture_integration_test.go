//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const iamIdentityDefaultRole = `
INSERT INTO iam_roles (
    id, scope_kind, scope_id, organization_id, name, name_key,
    description, status, is_protected, is_default, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, retired_at
) VALUES (
    'a1000000-0000-4000-8000-000000000001', 'tenant', $1, NULL,
    'Default Member', pg_catalog.convert_to('Default Member','UTF8'), NULL,
    'active', false, true, 'role',
    '41000000-0000-4000-8000-000000000003', '{}'::jsonb,
    pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')), 1,
    '2026-07-16T00:20:00Z', '2026-07-16T00:20:00Z', NULL
)`

func seedIAMIdentityRuntime(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		seedIAMRBACCatalog(t, ctx, database, tenant.schema)
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_users SET authorization_version=1, revision=revision+1,
    updated_at='2026-07-16T00:09:00Z'`)
		seedIAMIdentityProtectedPolicy(t, ctx, database, tenant)
		mustExecIAMTenant(t, ctx, database, tenant.schema,
			iamIdentityDefaultRole, tenant.id)
	}
}

type iamIdentityActor struct {
	tokenVersion         int64
	authorizationVersion int64
	revision             int64
}

func readIAMIdentityActor(
	t *testing.T, ctx context.Context, database *sql.DB, schema, userID string,
) iamIdentityActor {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	var actor iamIdentityActor
	if err := transaction.QueryRowContext(ctx, `
SELECT token_version,authorization_version,revision FROM iam_users
WHERE id=$1`, userID).Scan(
		&actor.tokenVersion, &actor.authorizationVersion, &actor.revision,
	); err != nil {
		t.Fatal(err)
	}
	if actor.tokenVersion < 1 || actor.authorizationVersion < 1 {
		t.Fatalf("invalid IAM identity actor versions: %+v", actor)
	}
	return actor
}

func withIAMIdentityRuntimeTx(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
	action func(*sql.Tx),
) {
	t.Helper()
	runIAMIdentityRuntimeTx(t, ctx, database, schema, true, action)
}

func probeIAMIdentityRuntimeTx(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
	action func(*sql.Tx),
) {
	t.Helper()
	runIAMIdentityRuntimeTx(t, ctx, database, schema, false, action)
}

func runIAMIdentityRuntimeTx(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
	commit bool, action func(*sql.Tx),
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx,
		"SET SESSION AUTHORIZATION harpy_v2_runtime"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := connection.ExecContext(context.Background(),
			"RESET SESSION AUTHORIZATION"); err != nil {
			t.Errorf("reset IAM identity runtime session: %v", err)
		}
	}()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		"pg_catalog, "+schema); err != nil {
		t.Fatal(err)
	}
	action(transaction)
	if !commit {
		if err := transaction.Rollback(); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
