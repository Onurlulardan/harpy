package main

import (
	"context"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type roleObservation struct {
	state    cutoverv1.AuthorityQuiescedRoleState
	verifier string
}

type quiescenceRun struct {
	policy     cutoverv1.AuthorityQuiescedRolePolicy
	pre        roleObservation
	temporary  roleObservation
	terminated int
}

func quiesceMigrator(
	ctx context.Context, connection *pgx.Conn,
	policy cutoverv1.AuthorityQuiescedRolePolicy,
) (quiescenceRun, error) {
	observed, err := readQuiescedRole(ctx, connection, policy.Principal)
	if err != nil {
		return quiescenceRun{}, err
	}
	run := quiescenceRun{policy: policy}
	switch observed.state {
	case policy.PreState:
		run.pre = observed
		transaction, beginErr := connection.Begin(ctx)
		if beginErr != nil {
			return quiescenceRun{}, fmt.Errorf("migrator quiescence transaction failed")
		}
		statement := "ALTER ROLE " + pgx.Identifier{policy.Principal}.Sanitize() +
			" NOLOGIN CONNECTION LIMIT 0"
		if _, err = transaction.Exec(ctx, statement); err == nil {
			err = transaction.Commit(ctx)
		} else {
			_ = transaction.Rollback(context.Background())
		}
		if err != nil {
			return quiescenceRun{}, fmt.Errorf("migrator quiescence failed")
		}
	case policy.TemporaryState:
		run.pre = roleObservation{state: policy.PreState, verifier: observed.verifier}
	default:
		return quiescenceRun{}, fmt.Errorf("migrator role state differs from signed policy")
	}
	run.temporary, err = readQuiescedRole(ctx, connection, policy.Principal)
	if err != nil || run.temporary.state != policy.TemporaryState ||
		run.temporary.verifier != run.pre.verifier {
		return quiescenceRun{}, fmt.Errorf("temporary migrator role state differs")
	}
	run.terminated, err = terminateRoleSessions(ctx, connection, policy.Principal)
	if err != nil {
		return quiescenceRun{}, err
	}
	return run, nil
}

func readQuiescedRole(
	ctx context.Context, query catalogQuerier, principal string,
) (roleObservation, error) {
	var value roleObservation
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
FROM pg_catalog.pg_authid AS role WHERE role.rolname=$1`, principal).Scan(
		&value.state.Login, &value.state.Inherit, &value.state.Superuser,
		&value.state.CreateDB, &value.state.CreateRole, &value.state.Replication,
		&value.state.BypassRLS, &value.state.ConnectionLimit, &value.state.PasswordState,
		&value.state.ValidUntil, &value.state.MembershipCount, &value.verifier,
	)
	if err != nil {
		return roleObservation{}, fmt.Errorf("migrator role observation failed")
	}
	return value, nil
}

func terminateRoleSessions(
	ctx context.Context, connection *pgx.Conn, principal string,
) (int, error) {
	total := 0
	for {
		var terminated, remaining int
		err := connection.QueryRow(ctx, `WITH targets AS (
    SELECT activity.pid FROM pg_catalog.pg_stat_activity AS activity
    WHERE activity.usename=$1 AND activity.pid<>pg_catalog.pg_backend_pid()
), terminated AS (
    SELECT pg_catalog.pg_terminate_backend(pid) AS stopped FROM targets
)
SELECT count(*) FILTER (WHERE stopped)::integer,
       (SELECT count(*)::integer FROM pg_catalog.pg_stat_activity
        WHERE usename=$1 AND pid<>pg_catalog.pg_backend_pid())
FROM terminated`, principal).Scan(&terminated, &remaining)
		if err != nil {
			return 0, fmt.Errorf("migrator session termination failed")
		}
		total += terminated
		if remaining == 0 {
			return total, nil
		}
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("migrator sessions did not quiesce")
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func restoreMigrator(
	ctx context.Context, query interface {
		Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	}, policy cutoverv1.AuthorityQuiescedRolePolicy,
) error {
	statement := "ALTER ROLE " + pgx.Identifier{policy.Principal}.Sanitize() +
		" LOGIN CONNECTION LIMIT -1"
	if _, err := query.Exec(ctx, statement); err != nil {
		return fmt.Errorf("migrator role restoration failed")
	}
	return nil
}
