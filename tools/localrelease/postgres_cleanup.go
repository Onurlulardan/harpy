package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var privilegedRoles = []string{"harpy_cluster_admin", "harpy_extension_admin"}

func cleanupPrivilegedRoles(ctx context.Context, admin *pgx.Conn) error {
	for _, role := range privilegedRoles {
		statement := "ALTER ROLE " + pgx.Identifier{role}.Sanitize() +
			" NOLOGIN PASSWORD NULL VALID UNTIL 'infinity'"
		if _, err := admin.Exec(ctx, statement); err != nil {
			return fmt.Errorf("disable privileged role %s: %w", role, err)
		}
		if _, err := admin.Exec(ctx, `SELECT pg_catalog.pg_terminate_backend(pid)
FROM pg_catalog.pg_stat_activity WHERE usename=$1 AND pid<>pg_catalog.pg_backend_pid()`,
			role); err != nil {
			return fmt.Errorf("terminate privileged role sessions: %w", err)
		}
		if err := revokeRoleMemberships(ctx, admin, role); err != nil {
			return err
		}
		if err := verifyCleanedRole(ctx, admin, role); err != nil {
			return err
		}
	}
	return nil
}

func revokeRoleMemberships(ctx context.Context, admin *pgx.Conn, principal string) error {
	rows, err := admin.Query(ctx, `SELECT role.rolname, member.rolname
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS role ON role.oid=membership.roleid
JOIN pg_catalog.pg_roles AS member ON member.oid=membership.member
WHERE role.rolname=$1 OR member.rolname=$1`, principal)
	if err != nil {
		return fmt.Errorf("list privileged memberships: %w", err)
	}
	memberships, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([2]string, error) {
		var value [2]string
		err := row.Scan(&value[0], &value[1])
		return value, err
	})
	if err != nil {
		return fmt.Errorf("collect privileged memberships: %w", err)
	}
	for _, membership := range memberships {
		statement := fmt.Sprintf("REVOKE %s FROM %s",
			pgx.Identifier{membership[0]}.Sanitize(), pgx.Identifier{membership[1]}.Sanitize())
		if _, err := admin.Exec(ctx, statement); err != nil {
			return fmt.Errorf("revoke privileged membership: %w", err)
		}
	}
	return nil
}

func verifyCleanedRole(ctx context.Context, admin *pgx.Conn, role string) error {
	var exact bool
	err := admin.QueryRow(ctx, `SELECT NOT subject.rolcanlogin
AND subject.rolpassword IS NULL AND subject.rolvaliduntil='infinity'::timestamptz
AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_stat_activity WHERE usename=$1)
AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
WHERE membership.roleid=subject.oid OR membership.member=subject.oid)
FROM pg_catalog.pg_authid AS subject WHERE subject.rolname=$1`, role).Scan(&exact)
	if err != nil || !exact {
		return fmt.Errorf("privileged role %s cleanup differs: %w", role, err)
	}
	return nil
}
