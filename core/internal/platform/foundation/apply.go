package foundation

import "context"

func Apply(ctx context.Context, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return failed("configuration", cfg, false, err)
	}
	if err := Preflight(ctx, cfg); err != nil {
		return err
	}
	admin, err := connect(ctx, cfg.AdminDSN)
	if err != nil {
		return failed("connect", cfg, true, err)
	}
	defer admin.Close(ctx)
	if err := ensureRoles(ctx, admin, cfg); err != nil {
		return err
	}
	if err := ensureDatabase(ctx, admin, cfg); err != nil {
		return err
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
	if err := ensurePlatformSchema(ctx, target, cfg); err != nil {
		return err
	}
	if err := ensureExtension(ctx, target, cfg); err != nil {
		return err
	}
	if err := ensurePrivileges(ctx, target, cfg); err != nil {
		return err
	}
	if err := ensureSessionDefaults(ctx, target, cfg); err != nil {
		return err
	}
	return assertNoApplicationRelations(ctx, target, cfg)
}
