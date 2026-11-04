package foundation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func Preflight(ctx context.Context, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return failed("configuration", cfg, false, err)
	}
	admin, err := connect(ctx, cfg.AdminDSN)
	if err != nil {
		return failed("connect", cfg, true, err)
	}
	defer admin.Close(ctx)
	if err := preflightRoles(ctx, admin, cfg); err != nil {
		return err
	}
	exists, owner, sessionUser, err := lookupDatabase(ctx, admin, cfg.Database)
	if err != nil {
		return failed("preflight", cfg, true, err)
	}
	if !exists {
		return nil
	}
	if owner != RoleOwner && owner != sessionUser {
		return failed("preflight", cfg, false, fmt.Errorf(
			"database %s is owned by %s", cfg.Database, owner,
		))
	}
	targetDSN, err := databaseDSN(cfg.AdminDSN, cfg.Database)
	if err != nil {
		return failed("configuration", cfg, false, err)
	}
	target, err := connect(ctx, targetDSN)
	if err != nil {
		return failed("connect", cfg, true, err)
	}
	defer target.Close(ctx)
	return preflightTarget(ctx, target, cfg)
}

func preflightRoles(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	rows, err := conn.Query(ctx, `SELECT rolname FROM pg_catalog.pg_roles
WHERE rolname LIKE 'harpy\_%' ESCAPE '\'`)
	if err != nil {
		return failed("preflight", cfg, true, err)
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return failed("preflight", cfg, true, err)
	}
	for _, name := range names {
		if !allowedRole(name) {
			return failed("preflight", cfg, false, fmt.Errorf(
				"unexpected role %s", name,
			))
		}
	}
	for _, spec := range managedRoles(cfg) {
		state, found, err := readRole(ctx, conn, spec.name)
		if err != nil {
			return failed("preflight", cfg, true, err)
		}
		if found && !state.matches(spec) {
			return failed("preflight", cfg, false, fmt.Errorf(
				"role %s has unexpected attributes", spec.name,
			))
		}
	}
	found, adminOption, inheritOption, setOption, err := readMembership(
		ctx, conn, RoleGoose, RoleOwner,
	)
	if err != nil {
		return failed("preflight", cfg, true, err)
	}
	if found && (adminOption || inheritOption || !setOption) {
		return failed("preflight", cfg, false, fmt.Errorf(
			"membership %s to %s has unexpected options", RoleGoose, RoleOwner,
		))
	}
	return nil
}

func preflightTarget(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	if err := preflightSchemas(ctx, conn, cfg); err != nil {
		return err
	}
	if err := preflightExtensions(ctx, conn, cfg); err != nil {
		return err
	}
	if err := preflightRelations(ctx, conn, cfg); err != nil {
		return err
	}
	return preflightSchemaPrivileges(ctx, conn, cfg)
}
