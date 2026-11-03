package foundation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func ensurePrivileges(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return failed("privilege", cfg, true, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	statements := []string{
		fmt.Sprintf("REVOKE ALL ON DATABASE %s FROM PUBLIC", ident(cfg.Database)),
		fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO %s, %s, %s",
			ident(cfg.Database), ident(RoleGoose), ident(RoleAPI), ident(RoleWorker)),
		"REVOKE CREATE ON SCHEMA public FROM PUBLIC",
		fmt.Sprintf("REVOKE ALL ON SCHEMA %s FROM PUBLIC", ident(PlatformSchema)),
		fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s, %s, %s",
			ident(PlatformSchema), ident(RoleGoose), ident(RoleAPI), ident(RoleWorker)),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %s, %s`,
			ident(RoleOwner), ident(PlatformSchema), ident(RoleAPI), ident(RoleWorker)),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s
GRANT USAGE, SELECT ON SEQUENCES TO %s, %s`,
			ident(RoleOwner), ident(PlatformSchema), ident(RoleAPI), ident(RoleWorker)),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s
GRANT EXECUTE ON FUNCTIONS TO %s, %s`,
			ident(RoleOwner), ident(PlatformSchema), ident(RoleAPI), ident(RoleWorker)),
	}
	for _, statement := range statements {
		if _, err := tx.Exec(ctx, statement); err != nil {
			return failed("privilege", cfg, true, fmt.Errorf("grant foundation privilege: %w", err))
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return failed("privilege", cfg, true, err)
	}
	return verifyRuntimePrivileges(ctx, conn, cfg)
}

func verifyRuntimePrivileges(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	for _, role := range []string{RoleAPI, RoleWorker, RoleGoose} {
		var connect, createDB, temporary, createSchema bool
		err := conn.QueryRow(ctx, `SELECT
pg_catalog.has_database_privilege($1, $2, 'CONNECT'),
pg_catalog.has_database_privilege($1, $2, 'CREATE'),
pg_catalog.has_database_privilege($1, $2, 'TEMP'),
pg_catalog.has_schema_privilege($1, $3, 'CREATE')`,
			role, cfg.Database, PlatformSchema).Scan(
			&connect, &createDB, &temporary, &createSchema,
		)
		if err != nil {
			return failed("privilege", cfg, true, err)
		}
		if !connect || createDB || temporary || createSchema {
			return failed("privilege", cfg, false, fmt.Errorf(
				"role %s privileges are not the runtime minimum", role,
			))
		}
	}
	return nil
}

func ensureSessionDefaults(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	statements := []string{
		fmt.Sprintf("ALTER DATABASE %s SET search_path TO pg_catalog", ident(cfg.Database)),
		fmt.Sprintf("ALTER ROLE %s IN DATABASE %s SET search_path TO pg_catalog",
			ident(RoleAPI), ident(cfg.Database)),
		fmt.Sprintf("ALTER ROLE %s IN DATABASE %s SET search_path TO pg_catalog",
			ident(RoleWorker), ident(cfg.Database)),
		fmt.Sprintf("ALTER ROLE %s IN DATABASE %s SET search_path TO %s, pg_catalog",
			ident(RoleGoose), ident(cfg.Database), ident(PlatformSchema)),
	}
	for _, statement := range statements {
		if _, err := conn.Exec(ctx, statement); err != nil {
			return failed("privilege", cfg, true, fmt.Errorf("set search path: %w", err))
		}
	}
	return nil
}
