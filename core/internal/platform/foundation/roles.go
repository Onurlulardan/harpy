package foundation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type roleSpec struct {
	name      string
	login     bool
	inherit   bool
	connLimit int
	password  string
}

type roleState struct {
	login       bool
	inherit     bool
	superuser   bool
	createDB    bool
	createRole  bool
	replication bool
	bypassRLS   bool
	connLimit   int
}

func managedRoles(cfg Config) []roleSpec {
	return []roleSpec{
		{name: RoleOwner, connLimit: -1},
		{name: RoleGoose, login: true, connLimit: 4, password: cfg.GoosePassword},
		{name: RoleAPI, login: true, inherit: true, connLimit: 20, password: cfg.APIPassword},
		{name: RoleWorker, login: true, inherit: true, connLimit: 20, password: cfg.WorkerPassword},
	}
}

func readRole(ctx context.Context, conn *pgx.Conn, name string) (roleState, bool, error) {
	var state roleState
	err := conn.QueryRow(ctx, `SELECT rolcanlogin, rolinherit, rolsuper, rolcreatedb,
rolcreaterole, rolreplication, rolbypassrls, rolconnlimit
FROM pg_catalog.pg_roles WHERE rolname = $1`, name).Scan(
		&state.login, &state.inherit, &state.superuser, &state.createDB,
		&state.createRole, &state.replication, &state.bypassRLS, &state.connLimit,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return roleState{}, false, nil
		}
		return roleState{}, false, fmt.Errorf("read role %s: %w", name, err)
	}
	return state, true, nil
}

func (state roleState) matches(spec roleSpec) bool {
	return state.login == spec.login && state.inherit == spec.inherit &&
		!state.superuser && !state.createDB && !state.createRole &&
		!state.replication && !state.bypassRLS && state.connLimit == spec.connLimit
}

func ensureRoles(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	for _, spec := range managedRoles(cfg) {
		if err := ensureRole(ctx, conn, cfg, spec); err != nil {
			return err
		}
	}
	return ensureOwnerMembership(ctx, conn, cfg)
}

func ensureRole(ctx context.Context, conn *pgx.Conn, cfg Config, spec roleSpec) error {
	state, found, err := readRole(ctx, conn, spec.name)
	if err != nil {
		return failed("role", cfg, true, err)
	}
	if found {
		if !state.matches(spec) {
			return failed("role", cfg, false, fmt.Errorf(
				"role %s has unexpected attributes", spec.name,
			))
		}
		return setRolePassword(ctx, conn, cfg, spec)
	}
	statement := fmt.Sprintf(
		"CREATE ROLE %s %s %s NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT %d PASSWORD NULL",
		ident(spec.name), loginClause(spec.login), inheritClause(spec.inherit), spec.connLimit,
	)
	if _, err := conn.Exec(ctx, statement); err != nil {
		return failed("role", cfg, true, fmt.Errorf("create role %s: %w", spec.name, err))
	}
	return setRolePassword(ctx, conn, cfg, spec)
}

func setRolePassword(ctx context.Context, conn *pgx.Conn, cfg Config, spec roleSpec) error {
	if spec.password == "" {
		return nil
	}
	if _, err := conn.Exec(ctx, "SET password_encryption = 'scram-sha-256'"); err != nil {
		return failed("role", cfg, true, err)
	}
	var statement string
	err := conn.QueryRow(ctx,
		"SELECT pg_catalog.format('ALTER ROLE %I PASSWORD %L', $1::text, $2::text)",
		spec.name, spec.password,
	).Scan(&statement)
	if err != nil {
		return failed("role", cfg, true, fmt.Errorf("quote password for %s: %w", spec.name, err))
	}
	if _, err := conn.Exec(ctx, statement); err != nil {
		return includeSecret(failed(
			"role", cfg, true, fmt.Errorf("set password for %s: %w", spec.name, err),
		), spec.password)
	}
	return nil
}

func ensureOwnerMembership(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	found, adminOption, inheritOption, setOption, err := readMembership(
		ctx, conn, RoleGoose, RoleOwner,
	)
	if err != nil {
		return failed("role", cfg, true, err)
	}
	if found {
		if adminOption || inheritOption || !setOption {
			return failed("role", cfg, false, fmt.Errorf(
				"membership %s to %s has unexpected options", RoleGoose, RoleOwner,
			))
		}
		return nil
	}
	statement := fmt.Sprintf(
		"GRANT %s TO %s WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		ident(RoleOwner), ident(RoleGoose),
	)
	if _, err := conn.Exec(ctx, statement); err != nil {
		return failed("role", cfg, true, fmt.Errorf("grant owner membership: %w", err))
	}
	return nil
}

func readMembership(
	ctx context.Context, conn *pgx.Conn, member, role string,
) (bool, bool, bool, bool, error) {
	var adminOption, inheritOption, setOption bool
	err := conn.QueryRow(ctx, `SELECT membership.admin_option, membership.inherit_option,
membership.set_option
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS member ON member.oid = membership.member
JOIN pg_catalog.pg_roles AS role ON role.oid = membership.roleid
WHERE member.rolname = $1 AND role.rolname = $2`, member, role).Scan(
		&adminOption, &inheritOption, &setOption,
	)
	if err == pgx.ErrNoRows {
		return false, false, false, false, nil
	}
	if err != nil {
		return false, false, false, false, fmt.Errorf("read membership %s: %w", member, err)
	}
	return true, adminOption, inheritOption, setOption, nil
}

func loginClause(login bool) string {
	if login {
		return "LOGIN"
	}
	return "NOLOGIN"
}

func inheritClause(inherit bool) string {
	if inherit {
		return "INHERIT"
	}
	return "NOINHERIT"
}
