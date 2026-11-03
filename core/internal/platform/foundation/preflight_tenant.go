package foundation

import (
	"context"
	"fmt"
)

func preflightTenant(ctx context.Context, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return failed("configuration", cfg, false, err)
	}
	admin, err := connect(ctx, cfg.AdminDSN)
	if err != nil {
		return failed("connect", cfg, true, err)
	}
	defer admin.Close(ctx)
	if err = preflightRoles(ctx, admin, cfg); err != nil {
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
	if err = preflightSchemas(ctx, target, cfg); err != nil {
		return err
	}
	if err = preflightExtensions(ctx, target, cfg); err != nil {
		return err
	}
	return preflightSchemaPrivileges(ctx, target, cfg)
}
