//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

type builtAuthorityRoleObservation struct {
	state    cutoverv1.AuthorityQuiescedRoleState
	verifier string
}

type builtAuthorityQuiescence struct {
	pre        builtAuthorityRoleObservation
	temporary  builtAuthorityRoleObservation
	post       builtAuthorityRoleObservation
	terminated int
}

func beginBuiltAuthorityQuiescence(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) builtAuthorityQuiescence {
	t.Helper()
	assertBuiltOwnerSetPrincipals(t, ctx, database.admin)
	pre := readBuiltAuthorityMigrator(t, ctx, database.admin)
	if _, err := database.admin.Exec(ctx, `ALTER ROLE harpy_migrator
NOLOGIN CONNECTION LIMIT 0`); err != nil {
		t.Fatalf("commit authority quiescence: %v", err)
	}
	temporary := readBuiltAuthorityMigrator(t, ctx, database.admin)
	terminated := terminateBuiltAuthorityMigratorSessions(t, ctx, database.admin)
	assertBuiltAuthorityMigratorSessions(t, ctx, database.admin, 0)
	if temporary.state != builtTemporaryMigratorState(pre.state) ||
		temporary.verifier != pre.verifier {
		t.Fatal("temporary migrator profile or password verifier differs")
	}
	return builtAuthorityQuiescence{
		pre: pre, temporary: temporary, terminated: terminated,
	}
}

func readBuiltAuthorityMigrator(
	t *testing.T, ctx context.Context, query builtAuthorityQuerier,
) builtAuthorityRoleObservation {
	t.Helper()
	var observation builtAuthorityRoleObservation
	err := query.QueryRow(ctx, `SELECT role.rolcanlogin, role.rolinherit, role.rolsuper,
    role.rolcreatedb, role.rolcreaterole, role.rolreplication, role.rolbypassrls,
    role.rolconnlimit,
    CASE WHEN role.rolpassword IS NULL THEN 'null'
         WHEN role.rolpassword LIKE 'SCRAM-SHA-256$%' THEN 'scram_sha_256'
         ELSE 'unsupported' END,
    CASE WHEN role.rolvaliduntil IS NULL OR role.rolvaliduntil='infinity'::timestamptz
         THEN 'infinity' ELSE role.rolvaliduntil::text END,
    (SELECT count(*) FROM pg_catalog.pg_auth_members AS membership
     WHERE membership.member=role.oid OR membership.roleid=role.oid),
    COALESCE(role.rolpassword,'')
FROM pg_catalog.pg_authid AS role WHERE role.rolname='harpy_migrator'`).Scan(
		&observation.state.Login, &observation.state.Inherit,
		&observation.state.Superuser, &observation.state.CreateDB,
		&observation.state.CreateRole, &observation.state.Replication,
		&observation.state.BypassRLS, &observation.state.ConnectionLimit,
		&observation.state.PasswordState, &observation.state.ValidUntil,
		&observation.state.MembershipCount, &observation.verifier,
	)
	if err != nil {
		t.Fatalf("read authority migrator profile: %v", err)
	}
	return observation
}

func builtTemporaryMigratorState(
	atRest cutoverv1.AuthorityQuiescedRoleState,
) cutoverv1.AuthorityQuiescedRoleState {
	state := atRest
	state.Login = false
	state.ConnectionLimit = 0
	return state
}

func restoreBuiltAuthorityMigrator(
	t *testing.T, ctx context.Context, transaction pgx.Tx,
	quiescence builtAuthorityQuiescence,
) builtAuthorityQuiescence {
	t.Helper()
	if _, err := transaction.Exec(ctx, `ALTER ROLE harpy_migrator
LOGIN CONNECTION LIMIT -1`); err != nil {
		t.Fatalf("restore authority migrator: %v", err)
	}
	quiescence.post = readBuiltAuthorityMigrator(t, ctx, transaction)
	if quiescence.post.state != quiescence.pre.state ||
		quiescence.post.verifier != quiescence.pre.verifier {
		t.Fatal("restored migrator profile or password verifier differs")
	}
	return quiescence
}

func terminateBuiltAuthorityMigratorSessions(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) int {
	t.Helper()
	total, deadline := 0, time.Now().Add(10*time.Second)
	for time.Now().Before(deadline) {
		var terminated int
		err := admin.QueryRow(ctx, `SELECT count(*) FILTER (
    WHERE pg_catalog.pg_terminate_backend(activity.pid))::integer
FROM pg_catalog.pg_stat_activity AS activity
WHERE activity.usename='harpy_migrator'
  AND activity.pid<>pg_catalog.pg_backend_pid()`).Scan(&terminated)
		if err != nil {
			t.Fatalf("terminate authority migrator sessions: %v", err)
		}
		total += terminated
		if builtAuthorityMigratorSessionCount(t, ctx, admin) == 0 {
			return total
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("authority migrator sessions did not quiesce")
	return 0
}

func assertBuiltAuthorityMigratorSessions(
	t *testing.T, ctx context.Context, query builtAuthorityQuerier, expected int,
) {
	t.Helper()
	if count := builtAuthorityMigratorSessionCount(t, ctx, query); count != expected {
		t.Fatalf("authority migrator sessions=%d want=%d", count, expected)
	}
}

func builtAuthorityMigratorSessionCount(
	t *testing.T, ctx context.Context, query builtAuthorityQuerier,
) int {
	t.Helper()
	var count int
	if err := query.QueryRow(ctx, `SELECT count(*) FROM pg_catalog.pg_stat_activity
WHERE usename='harpy_migrator'`).Scan(&count); err != nil {
		t.Fatalf("count authority migrator sessions: %v", err)
	}
	return count
}

func assertBuiltOwnerSetPrincipals(
	t *testing.T, ctx context.Context, query builtAuthorityQuerier,
) {
	t.Helper()
	var exact bool
	err := query.QueryRow(ctx, `WITH RECURSIVE capable(role_oid) AS (
    SELECT membership.member FROM pg_catalog.pg_auth_members AS membership
    JOIN pg_catalog.pg_roles AS role ON role.oid=membership.roleid
    WHERE role.rolname='harpy_owner' AND membership.set_option
    UNION
    SELECT membership.member FROM pg_catalog.pg_auth_members AS membership
    JOIN capable ON capable.role_oid=membership.roleid WHERE membership.set_option
)
SELECT COALESCE(pg_catalog.array_agg(role.rolname ORDER BY role.rolname),ARRAY[]::text[])
       = ARRAY['harpy_migrator']::text[]
FROM capable JOIN pg_catalog.pg_roles AS role ON role.oid=capable.role_oid`).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("harpy_owner SET principal graph differs: exact=%t err=%v", exact, err)
	}
}
