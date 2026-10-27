//go:build integration

package migrator

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertProductionRoleFixture(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
) {
	t.Helper()
	expected := builtEntrypointRolesManifest().Roles
	var count int
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM pg_catalog.pg_roles
WHERE rolname LIKE 'harpy_%'`).Scan(&count); err != nil || count != len(expected) {
		t.Fatalf("production role fixture count=%d want=%d err=%v", count, len(expected), err)
	}
	for _, role := range expected {
		var login, inherit, superuser, createDB, createRole, replication, bypassRLS bool
		var passwordNull, validUntilInfinity bool
		var connectionLimit int
		err := admin.QueryRow(ctx, `SELECT role.rolcanlogin, role.rolinherit,
role.rolsuper, role.rolcreatedb, role.rolcreaterole, role.rolreplication,
role.rolbypassrls, role.rolconnlimit, role.rolpassword IS NULL,
role.rolvaliduntil='infinity'::timestamptz
FROM pg_catalog.pg_authid AS role WHERE role.rolname=$1`, role.Name).Scan(
			&login, &inherit, &superuser, &createDB, &createRole,
			&replication, &bypassRLS, &connectionLimit,
			&passwordNull, &validUntilInfinity,
		)
		if err != nil || login != role.Login || inherit != role.Inherit ||
			superuser != role.Superuser || createDB != role.CreateDB ||
			createRole != role.CreateRole || replication != role.Replication ||
			bypassRLS != role.BypassRLS || connectionLimit != role.ConnectionLimit ||
			!passwordNull || !validUntilInfinity {
			t.Fatalf("production role %s differs: err=%v", role.Name, err)
		}
	}
	assertProductionRoleMembership(
		t, ctx, admin, "harpy_v2_runtime", "harpy_runtime",
	)
	assertProductionRoleMembership(
		t, ctx, admin, "harpy_migrator", "harpy_owner",
	)
	assertProductionRoleMembership(
		t, ctx, admin, "harpy_backup", "harpy_backup_reader",
	)
	assertProductionRoleMembership(
		t, ctx, admin, "harpy_restore", "harpy_owner",
	)
	assertProductionAuthorityOwnerBoundary(t, ctx, admin)
}

func assertProductionAuthorityOwnerBoundary(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	var exact bool
	err := admin.QueryRow(ctx, `SELECT count(*)=1
FROM pg_catalog.pg_authid AS role
WHERE role.rolname='harpy_cutover_authority_owner'
  AND role.rolpassword IS NULL AND role.rolvaliduntil='infinity'::timestamptz
  AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
      WHERE membership.roleid=role.oid OR membership.member=role.oid)`).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("production authority owner boundary exact=%t err=%v", exact, err)
	}
}

func assertProductionRoleMembership(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
	member string,
	role string,
) {
	t.Helper()
	var adminOption, inheritOption, setOption bool
	err := admin.QueryRow(ctx, `SELECT membership.admin_option,
membership.inherit_option, membership.set_option
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS member ON member.oid=membership.member
JOIN pg_catalog.pg_roles AS role ON role.oid=membership.roleid
WHERE member.rolname=$1 AND role.rolname=$2`, member, role).Scan(
		&adminOption, &inheritOption, &setOption,
	)
	if err != nil || adminOption || inheritOption || !setOption {
		t.Fatalf("production membership %s->%s differs: %v", member, role, err)
	}
}
