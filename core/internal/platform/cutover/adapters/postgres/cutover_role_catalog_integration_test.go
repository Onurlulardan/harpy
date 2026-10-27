//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertCutoverRoleCatalogBoundary(
	t *testing.T, ctx context.Context, admin *pgx.Conn, name string,
) {
	t.Helper()
	var inherit, superuser, createDB, createRole, bypassRLS bool
	var memberships, ownerships int
	err := admin.QueryRow(ctx, `SELECT role.rolinherit, role.rolsuper,
    role.rolcreatedb, role.rolcreaterole, role.rolbypassrls,
    (SELECT count(*) FROM pg_catalog.pg_auth_members AS membership
     WHERE membership.member=role.oid OR membership.roleid=role.oid),
    (SELECT count(*) FROM pg_catalog.pg_class AS object WHERE object.relowner=role.oid) +
    (SELECT count(*) FROM pg_catalog.pg_proc AS object WHERE object.proowner=role.oid) +
    (SELECT count(*) FROM pg_catalog.pg_namespace AS object WHERE object.nspowner=role.oid) +
    (SELECT count(*) FROM pg_catalog.pg_database AS object WHERE object.datdba=role.oid) +
    (SELECT count(*) FROM pg_catalog.pg_type AS object WHERE object.typowner=role.oid)
FROM pg_catalog.pg_authid AS role WHERE role.rolname=$1`, name).Scan(
		&inherit, &superuser, &createDB, &createRole, &bypassRLS,
		&memberships, &ownerships,
	)
	if err != nil || inherit || superuser || createDB || createRole || bypassRLS ||
		memberships != 0 || ownerships != 0 {
		t.Fatalf("role %s authority = %t %t %t %t %t memberships=%d ownerships=%d err=%v",
			name, inherit, superuser, createDB, createRole, bypassRLS,
			memberships, ownerships, err)
	}
}

func verifyDecoderPrivilegeBoundary(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	decoder := connectCutoverRole(t, ctx, database.dsn, "harpy_cutover_decoder")
	defer decoder.Close(ctx)
	var rows int
	if err := decoder.QueryRow(ctx,
		"SELECT count(*) FROM legacy_source.allowed_rows").Scan(&rows); err != nil || rows != 1 {
		t.Fatalf("decoder allowlisted SELECT rows=%d err=%v", rows, err)
	}
	denied := []string{
		`INSERT INTO legacy_source.allowed_rows VALUES ('80000000-0000-4000-8000-000000000002')`,
		`SELECT count(*) FROM platform.integration_business_rows`,
		`CREATE TABLE legacy_source.decoder_forbidden_ddl (id uuid)`,
		`CREATE TEMPORARY TABLE decoder_forbidden_temp (id uuid)`,
		`SET ROLE harpy_owner`,
		`SET ROLE harpy_adopter`,
	}
	for _, statement := range denied {
		_, err := decoder.Exec(ctx, statement)
		assertPermissionDenied(t, err)
	}
	assertNoCutoverMembership(t, ctx, database, "harpy_cutover_decoder")
}
