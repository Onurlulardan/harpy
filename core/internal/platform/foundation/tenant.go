package foundation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type TenantNamespace struct {
	Schema string
	Role   string
}

func PrepareTenantNamespace(
	ctx context.Context, cfg Config, tenantID [16]byte, password string,
) (TenantNamespace, error) {
	if err := cfg.Validate(); err != nil {
		return TenantNamespace{}, failed("configuration", cfg, false, err)
	}
	if !secretOK(password) {
		return TenantNamespace{}, failed("tenant", cfg, false, fmt.Errorf(
			"tenant role password is missing or too short",
		))
	}
	schemaName, err := TenantSchemaName(tenantID)
	if err != nil {
		return TenantNamespace{}, failed("tenant", cfg, false, err)
	}
	roleName, err := TenantRoleName(tenantID)
	if err != nil {
		return TenantNamespace{}, failed("tenant", cfg, false, err)
	}
	namespace := TenantNamespace{Schema: schemaName, Role: roleName}
	if err := preflightTenant(ctx, cfg); err != nil {
		return TenantNamespace{}, err
	}
	targetDSN, err := databaseDSN(cfg.AdminDSN, cfg.Database)
	if err != nil {
		return TenantNamespace{}, failed("configuration", cfg, false, err)
	}
	target, err := connect(ctx, targetDSN)
	if err != nil {
		return TenantNamespace{}, failed("connect", cfg, true, err)
	}
	defer target.Close(ctx)
	spec := roleSpec{name: roleName, login: true, connLimit: 4, password: password}
	if err := ensureRole(ctx, target, cfg, spec); err != nil {
		return TenantNamespace{}, err
	}
	if err := ensureTenantSchema(ctx, target, cfg, namespace); err != nil {
		return TenantNamespace{}, err
	}
	if err := ensureTenantPrivileges(ctx, target, cfg, namespace); err != nil {
		return TenantNamespace{}, err
	}
	statement := fmt.Sprintf(
		"ALTER ROLE %s IN DATABASE %s SET search_path TO %s, pg_catalog",
		ident(roleName), ident(cfg.Database), ident(schemaName),
	)
	if _, err := target.Exec(ctx, statement); err != nil {
		return TenantNamespace{}, failed("tenant", cfg, true, err)
	}
	return namespace, nil
}

func ensureTenantSchema(
	ctx context.Context, conn *pgx.Conn, cfg Config, namespace TenantNamespace,
) error {
	var owner string
	err := conn.QueryRow(ctx, `SELECT pg_catalog.pg_get_userbyid(nspowner)
FROM pg_catalog.pg_namespace WHERE nspname = $1`, namespace.Schema).Scan(&owner)
	if err == pgx.ErrNoRows {
		statement := fmt.Sprintf(
			"CREATE SCHEMA %s AUTHORIZATION %s", ident(namespace.Schema), ident(namespace.Role),
		)
		if _, err := conn.Exec(ctx, statement); err != nil {
			return failed("tenant", cfg, true, fmt.Errorf("create tenant schema: %w", err))
		}
		return nil
	}
	if err != nil {
		return failed("tenant", cfg, true, err)
	}
	if owner != namespace.Role {
		return failed("tenant", cfg, false, fmt.Errorf(
			"schema %s is owned by %s", namespace.Schema, owner,
		))
	}
	return nil
}

func ensureTenantPrivileges(
	ctx context.Context, conn *pgx.Conn, cfg Config, namespace TenantNamespace,
) error {
	statements := []string{
		fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO %s",
			ident(cfg.Database), ident(namespace.Role)),
		fmt.Sprintf("REVOKE ALL ON SCHEMA %s FROM PUBLIC", ident(namespace.Schema)),
		fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s, %s",
			ident(namespace.Schema), ident(RoleAPI), ident(RoleWorker)),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %s, %s`,
			ident(namespace.Role), ident(namespace.Schema), ident(RoleAPI), ident(RoleWorker)),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s
GRANT USAGE, SELECT ON SEQUENCES TO %s, %s`,
			ident(namespace.Role), ident(namespace.Schema), ident(RoleAPI), ident(RoleWorker)),
	}
	for _, statement := range statements {
		if _, err := conn.Exec(ctx, statement); err != nil {
			return failed("tenant", cfg, true, fmt.Errorf("grant tenant privilege: %w", err))
		}
	}
	return nil
}
