package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/jackc/pgx/v5"
)

func reconcileRoles(ctx context.Context, admin *pgx.Conn) error {
	manifest := rolesManifest()
	for _, role := range manifest.Roles {
		var exists bool
		if err := admin.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_roles WHERE rolname=$1)",
			role.Name,
		).Scan(&exists); err != nil {
			return fmt.Errorf("inspect role %s: %w", role.Name, err)
		}
		verb := "CREATE ROLE"
		if exists {
			verb = "ALTER ROLE"
		}
		if _, err := admin.Exec(ctx, roleStatement(verb, role)); err != nil {
			return fmt.Errorf("reconcile role %s: %w", role.Name, err)
		}
	}
	if err := verifyManagedRoleSet(ctx, admin, manifest); err != nil {
		return err
	}
	for _, membership := range privilegesManifest().Memberships {
		statement := fmt.Sprintf(
			"GRANT %s TO %s WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
			pgx.Identifier{membership.Role}.Sanitize(),
			pgx.Identifier{membership.Member}.Sanitize(),
		)
		if _, err := admin.Exec(ctx, statement); err != nil {
			return fmt.Errorf("reconcile membership %s -> %s: %w",
				membership.Member, membership.Role, err)
		}
	}
	return verifyRequiredMemberships(ctx, admin)
}

func roleStatement(verb string, role postgresv1.PostgresRole) string {
	options := []string{
		booleanOption(role.Login, "LOGIN", "NOLOGIN"),
		booleanOption(role.Inherit, "INHERIT", "NOINHERIT"),
		booleanOption(role.Superuser, "SUPERUSER", "NOSUPERUSER"),
		booleanOption(role.CreateDB, "CREATEDB", "NOCREATEDB"),
		booleanOption(role.CreateRole, "CREATEROLE", "NOCREATEROLE"),
		booleanOption(role.Replication, "REPLICATION", "NOREPLICATION"),
		booleanOption(role.BypassRLS, "BYPASSRLS", "NOBYPASSRLS"),
		fmt.Sprintf("CONNECTION LIMIT %d", role.ConnectionLimit),
		"PASSWORD NULL", "VALID UNTIL 'infinity'",
	}
	return verb + " " + pgx.Identifier{role.Name}.Sanitize() + " " + strings.Join(options, " ")
}

func booleanOption(value bool, enabled, disabled string) string {
	if value {
		return enabled
	}
	return disabled
}

func verifyManagedRoleSet(
	ctx context.Context, admin *pgx.Conn, manifest postgresv1.RolesManifest,
) error {
	rows, err := admin.Query(ctx,
		`SELECT rolname FROM pg_catalog.pg_roles
WHERE rolname LIKE 'harpy\_%' ESCAPE '\' OR rolname='forge_runtime'`,
	)
	if err != nil {
		return fmt.Errorf("list managed roles: %w", err)
	}
	actual, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return fmt.Errorf("collect managed roles: %w", err)
	}
	expected := make([]string, len(manifest.Roles))
	for index, role := range manifest.Roles {
		expected[index] = role.Name
	}
	sort.Strings(actual)
	sort.Strings(expected)
	if strings.Join(actual, "\x00") != strings.Join(expected, "\x00") {
		return fmt.Errorf("managed role set differs: got %v want %v", actual, expected)
	}
	return nil
}

func verifyRequiredMemberships(ctx context.Context, admin *pgx.Conn) error {
	for _, membership := range privilegesManifest().Memberships {
		var adminOption, inheritOption, setOption bool
		err := admin.QueryRow(ctx, `SELECT membership.admin_option, membership.inherit_option,
membership.set_option FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS member ON member.oid=membership.member
JOIN pg_catalog.pg_roles AS role ON role.oid=membership.roleid
WHERE member.rolname=$1 AND role.rolname=$2`, membership.Member, membership.Role).Scan(
			&adminOption, &inheritOption, &setOption,
		)
		if err != nil || adminOption || inheritOption || !setOption {
			return fmt.Errorf("membership %s -> %s differs: %w",
				membership.Member, membership.Role, err)
		}
	}
	return nil
}

func setRolePassword(ctx context.Context, admin *pgx.Conn, role, password string) error {
	if _, err := admin.Exec(ctx, "SET password_encryption = 'scram-sha-256'"); err != nil {
		return err
	}
	var statement string
	if err := admin.QueryRow(ctx,
		"SELECT pg_catalog.format('ALTER ROLE %I PASSWORD %L', $1::text, $2::text)",
		role, password,
	).Scan(&statement); err != nil {
		return fmt.Errorf("quote %s password: %w", role, err)
	}
	if _, err := admin.Exec(ctx, statement); err != nil {
		return fmt.Errorf("set %s password: %w", role, err)
	}
	return nil
}
