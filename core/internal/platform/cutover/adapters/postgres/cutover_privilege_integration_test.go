//go:build integration

package postgres

import (
	"context"
	"testing"
)

func verifyCutoverRolePrivilegeBoundary(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase, role string,
) {
	t.Helper()
	connection := connectCutoverRole(t, ctx, database.dsn, role)
	defer connection.Close(ctx)
	statements := []string{
		`INSERT INTO platform.integration_business_rows
            VALUES ('90000000-0000-4000-8000-000000000001'::uuid)`,
		`CREATE TABLE platform.cutover_forbidden_ddl (id uuid)`,
		`CREATE TEMPORARY TABLE cutover_forbidden_temp (id uuid)`,
		`SET ROLE harpy_owner`,
	}
	if role == "harpy_adopter" {
		statements = append(statements, `SET ROLE harpy_cutover_finalizer`)
	} else {
		statements = append(statements,
			`SET ROLE harpy_adopter`,
			`SELECT * FROM platform.cutover_prepare_batch_v1(
                NULL::uuid,NULL::uuid,NULL::text,NULL::uuid,NULL::smallint,
                NULL::text,NULL::integer,NULL::pg_lsn,NULL::xid8,NULL::text,NULL::text)`,
		)
	}
	for _, statement := range statements {
		_, err := connection.Exec(ctx, statement)
		assertPermissionDenied(t, err)
	}
	assertNoCutoverMembership(t, ctx, database, role)
}

func assertNoCutoverMembership(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase, role string,
) {
	t.Helper()
	var memberships int
	err := database.admin.QueryRow(ctx, `SELECT count(*)
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS role
  ON role.oid = membership.member OR role.oid = membership.roleid
WHERE role.rolname = $1`, role).Scan(&memberships)
	if err != nil || memberships != 0 {
		t.Fatalf("role %s memberships = %d, err=%v", role, memberships, err)
	}
}
