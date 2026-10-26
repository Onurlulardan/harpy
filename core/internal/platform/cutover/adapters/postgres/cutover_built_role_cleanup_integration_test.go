//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func cleanupBuiltAdoptionRoles(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `
SELECT pg_catalog.pg_terminate_backend(activity.pid)
FROM pg_catalog.pg_stat_activity AS activity
WHERE activity.usename IN ('harpy_cutover_decoder','harpy_adopter')
  AND activity.pid <> pg_catalog.pg_backend_pid();
REVOKE SELECT ON legacy_source.allowed_rows FROM harpy_cutover_decoder;
ALTER ROLE harpy_cutover_decoder NOLOGIN NOREPLICATION CONNECTION LIMIT 0
PASSWORD NULL VALID UNTIL 'infinity';
ALTER ROLE harpy_adopter NOLOGIN NOREPLICATION CONNECTION LIMIT 0
PASSWORD NULL VALID UNTIL 'infinity';`)
	if err != nil {
		t.Fatalf("cleanup built adoption roles: %v", err)
	}
	assertBuiltCleanRole(t, ctx, database.admin, "harpy_cutover_decoder")
	assertBuiltCleanRole(t, ctx, database.admin, "harpy_adopter")
	assertBuiltRoleConnectionDenied(t, ctx, database, "harpy_cutover_decoder", builtCutoverPassword)
	assertBuiltRoleConnectionDenied(t, ctx, database, "harpy_adopter", builtCutoverPassword)
}

func cleanupBuiltFinalizerRole(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `
SELECT pg_catalog.pg_terminate_backend(activity.pid)
FROM pg_catalog.pg_stat_activity AS activity
WHERE activity.usename='harpy_cutover_finalizer'
  AND activity.pid <> pg_catalog.pg_backend_pid();
ALTER ROLE harpy_cutover_finalizer NOLOGIN NOREPLICATION CONNECTION LIMIT 0
PASSWORD NULL VALID UNTIL 'infinity';`)
	if err != nil {
		t.Fatalf("cleanup built finalizer role: %v", err)
	}
	assertBuiltCleanRole(t, ctx, database.admin, "harpy_cutover_finalizer")
	assertBuiltRoleConnectionDenied(t, ctx, database, "harpy_cutover_finalizer", builtCutoverPassword)
}

func assertBuiltCleanRole(
	t *testing.T, ctx context.Context, admin *pgx.Conn, role string,
) {
	t.Helper()
	var exact bool
	err := admin.QueryRow(ctx, `SELECT count(*)=1 FROM pg_catalog.pg_authid AS role
WHERE role.rolname=$1 AND NOT role.rolcanlogin AND NOT role.rolinherit
  AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
  AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=0
  AND role.rolpassword IS NULL
  AND role.rolvaliduntil = 'infinity'::timestamptz
  AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_stat_activity AS activity
                  WHERE activity.usename=role.rolname)
  AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                  WHERE membership.roleid=role.oid OR membership.member=role.oid)`, role).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("clean role %s differs: exact=%t err=%v", role, exact, err)
	}
}

func assertBuiltSealedRole(
	t *testing.T, ctx context.Context, admin *pgx.Conn, role string, expiresAt time.Time,
) {
	t.Helper()
	var exact bool
	err := admin.QueryRow(ctx, `SELECT count(*)=1 FROM pg_catalog.pg_authid AS role
WHERE role.rolname=$1 AND NOT role.rolcanlogin AND NOT role.rolinherit
  AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
  AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=0
  AND role.rolpassword IS NULL
  AND role.rolvaliduntil IS NOT DISTINCT FROM $2::timestamptz
  AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                  WHERE membership.roleid=role.oid OR membership.member=role.oid)`,
		role, expiresAt).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("sealed role %s differs: exact=%t err=%v", role, exact, err)
	}
}

func assertExpiredBuiltRolePasswordsDenied(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	past := time.Now().UTC().Add(-time.Second).Format(time.RFC3339Nano)
	roles := []struct {
		name        string
		replication string
	}{
		{name: "harpy_cutover_decoder", replication: "REPLICATION"},
		{name: "harpy_adopter", replication: "NOREPLICATION"},
		{name: "harpy_cutover_finalizer", replication: "NOREPLICATION"},
	}
	for _, role := range roles {
		statement := fmt.Sprintf(`ALTER ROLE %s LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE %s NOBYPASSRLS CONNECTION LIMIT 1 PASSWORD '%s' VALID UNTIL '%s'`,
			role.name, role.replication, builtCutoverExpiredPassword, past)
		if _, err := database.admin.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
		dsn := builtCutoverRoleDSN(t, database.dsn, role.name, builtCutoverExpiredPassword)
		connection, err := pgx.Connect(ctx, dsn)
		if err == nil {
			_ = connection.Close(ctx)
			t.Fatalf("expired %s password opened a new connection", role.name)
		}
		if role.name == "harpy_cutover_finalizer" {
			cleanupBuiltFinalizerRole(t, ctx, database)
		} else {
			cleanupBuiltAdoptionRoles(t, ctx, database)
		}
	}
}

func assertBuiltRoleConnectionDenied(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	role, password string,
) {
	t.Helper()
	dsn := builtCutoverRoleDSN(t, database.dsn, role, password)
	connection, err := pgx.Connect(ctx, dsn)
	if err == nil {
		_ = connection.Close(ctx)
		t.Fatalf("sealed %s credential opened a new connection", role)
	}
}

func assertBuiltRoleConnectionAccepted(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase, role string,
) {
	t.Helper()
	dsn := builtCutoverRoleDSN(t, database.dsn, role, builtCutoverPassword)
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("leased %s SCRAM connection failed: %v", role, err)
	}
	defer connection.Close(ctx)
	var sessionUser string
	if err := connection.QueryRow(ctx, "SELECT SESSION_USER").Scan(&sessionUser); err != nil || sessionUser != role {
		t.Fatalf("leased %s SCRAM session user=%s err=%v", role, sessionUser, err)
	}
}

func assertBuiltLeaseCatalog(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	role string, expiresAt time.Time, replication bool, limit int,
) {
	t.Helper()
	var exact bool
	err := admin.QueryRow(ctx, `SELECT count(*)=1 FROM pg_catalog.pg_authid AS role
WHERE role.rolname=$1 AND role.rolcanlogin AND NOT role.rolinherit
  AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
  AND role.rolreplication=$3 AND NOT role.rolbypassrls AND role.rolconnlimit=$4
  AND role.rolpassword LIKE 'SCRAM-SHA-256$%'
  AND role.rolvaliduntil IS NOT DISTINCT FROM $2::timestamptz`,
		role, expiresAt, replication, limit).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("lease catalog %s differs: exact=%t err=%v", role, exact, err)
	}
}
