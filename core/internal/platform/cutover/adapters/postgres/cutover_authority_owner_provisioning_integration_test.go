//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	cutoverAuthorityOwner = "harpy_cutover_authority_owner"
	clusterAdminPassword  = "built-cluster-admin-secret"
)

func TestCutoverAuthorityOwnerRequiresExternalPG18Provisioning(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startCutoverIntegrationDatabase(t, ctx)
	dropAuthorityOwnerIfPresent(t, ctx, database.admin)
	if externalAuthorityOwnerExact(t, ctx, database.admin) {
		t.Fatal("missing authority owner was implicitly created")
	}
	assertBuiltRoleAbsent(t, ctx, database.admin, cutoverAuthorityOwner)

	createExternalAuthorityOwner(t, ctx, database.admin)
	if !externalAuthorityOwnerExact(t, ctx, database.admin) {
		t.Fatal("external superuser provisioned authority owner differs")
	}
	clusterAdmin := connectBuiltClusterAdmin(t, ctx, database)
	assertClusterAdminCannotGrantAuthority(t, ctx, clusterAdmin)
	prepareProtectedAuthorityProbe(t, ctx, database.admin)
	assertAuthorityProbeExecuteGrant(t, ctx, clusterAdmin)
	assertClusterAdminCannotMutateAuthority(t, ctx, clusterAdmin)
	assertClusterAdminCannotChangeAuthorityACL(t, ctx, database.admin, clusterAdmin)
	if _, err := database.admin.Exec(ctx,
		"DROP SCHEMA cutover_authority CASCADE; DROP ROLE "+cutoverAuthorityOwner); err != nil {
		t.Fatal(err)
	}

	if _, err := clusterAdmin.Exec(ctx, exactAuthorityOwnerCreateSQL()); err != nil {
		t.Fatalf("cluster admin create authority owner: %v", err)
	}
	if externalAuthorityOwnerExact(t, ctx, database.admin) {
		t.Fatal("cluster-admin-created authority owner passed external preflight")
	}
	assertImplicitAuthorityAdminMembership(t, ctx, database.admin)
	if err := clusterAdmin.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := database.admin.Exec(ctx, `DROP ROLE harpy_cutover_authority_owner;
ALTER ROLE harpy_cluster_admin NOLOGIN CONNECTION LIMIT 0 PASSWORD NULL
VALID UNTIL 'infinity';`); err != nil {
		t.Fatal(err)
	}
}

func dropAuthorityOwnerIfPresent(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	var exists bool
	if err := admin.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname=$1)",
		cutoverAuthorityOwner).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		return
	}
	_, err := admin.Exec(ctx, `DROP SCHEMA IF EXISTS cutover_authority CASCADE;
REASSIGN OWNED BY harpy_cutover_authority_owner TO postgres;
DROP OWNED BY harpy_cutover_authority_owner;
DROP ROLE harpy_cutover_authority_owner;`)
	if err != nil {
		t.Fatalf("reset authority owner fixture: %v", err)
	}
}

func createExternalAuthorityOwner(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	if _, err := admin.Exec(ctx, exactAuthorityOwnerCreateSQL()); err != nil {
		t.Fatalf("external create authority owner: %v", err)
	}
}

func exactAuthorityOwnerCreateSQL() string {
	return `CREATE ROLE harpy_cutover_authority_owner NOLOGIN NOINHERIT
NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS
CONNECTION LIMIT -1 PASSWORD NULL VALID UNTIL 'infinity'`
}

func externalAuthorityOwnerExact(t *testing.T, ctx context.Context, admin *pgx.Conn) bool {
	t.Helper()
	var exact bool
	err := admin.QueryRow(ctx, `SELECT count(*)=1 FROM pg_catalog.pg_authid AS role
WHERE role.rolname=$1 AND NOT role.rolcanlogin AND NOT role.rolinherit
  AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
  AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=-1
  AND role.rolpassword IS NULL AND role.rolvaliduntil='infinity'::timestamptz
  AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                  WHERE membership.roleid=role.oid OR membership.member=role.oid)`,
		cutoverAuthorityOwner).Scan(&exact)
	if err != nil {
		t.Fatal(err)
	}
	return exact
}

func connectBuiltClusterAdmin(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) *pgx.Conn {
	t.Helper()
	expiresAt := time.Now().UTC().Add(time.Minute).Truncate(time.Microsecond)
	statement := fmt.Sprintf(`ALTER ROLE harpy_cluster_admin LOGIN CONNECTION LIMIT 1
PASSWORD '%s' VALID UNTIL '%s'`, clusterAdminPassword, expiresAt.Format(time.RFC3339Nano))
	if _, err := database.admin.Exec(ctx, statement); err != nil {
		t.Fatal(err)
	}
	dsn := builtCutoverRoleDSN(t, database.dsn, "harpy_cluster_admin", clusterAdminPassword)
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect leased cluster admin: %v", err)
	}
	return connection
}

func assertClusterAdminCannotGrantAuthority(
	t *testing.T, ctx context.Context, clusterAdmin *pgx.Conn,
) {
	t.Helper()
	_, err := clusterAdmin.Exec(ctx,
		"GRANT harpy_cutover_authority_owner TO harpy_cluster_admin")
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
		t.Fatalf("cluster admin authority self-grant error=%v", err)
	}
}

func assertImplicitAuthorityAdminMembership(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	var count int
	var adminOption bool
	err := admin.QueryRow(ctx, `SELECT count(*), bool_and(membership.admin_option)
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS role ON role.oid=membership.roleid
JOIN pg_catalog.pg_roles AS member ON member.oid=membership.member
WHERE role.rolname=$1 AND member.rolname='harpy_cluster_admin'`,
		cutoverAuthorityOwner).Scan(&count, &adminOption)
	if err != nil || count != 1 || !adminOption {
		t.Fatalf("implicit authority membership count=%d admin=%t err=%v", count, adminOption, err)
	}
}

func assertBuiltRoleAbsent(t *testing.T, ctx context.Context, admin *pgx.Conn, role string) {
	t.Helper()
	var count int
	if err := admin.QueryRow(ctx,
		"SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=$1", role).Scan(&count); err != nil || count != 0 {
		t.Fatalf("role %s count=%d err=%v", role, count, err)
	}
}
