package goose

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	"github.com/jackc/pgx/v5"
)

var tenantProvisioningRole = regexp.MustCompile(`^harpy_td_[0-9a-f]{32}$`)

func grantActivationRoutines(ctx context.Context, cfg foundation.Config, role string) error {
	return execRoleSQL(ctx, cfg, role, activationGrantSQL, "grant activation routine to tenant role")
}

func revokeActivationGrantOption(ctx context.Context, cfg foundation.Config, role string) error {
	return execRoleSQL(ctx, cfg, role, activationGrantOptionRevokeSQL, "revoke activation grant option")
}

func execRoleSQL(
	ctx context.Context,
	cfg foundation.Config,
	role string,
	statements func(string) ([]string, error),
	action string,
) error {
	sqlText, err := statements(role)
	if err != nil {
		return err
	}
	dsn, err := adminDatabaseDSN(cfg)
	if err != nil {
		return err
	}
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect activation grant: %w", err)
	}
	defer connection.Close(ctx)
	for _, statement := range sqlText {
		if _, err = connection.Exec(ctx, statement); err != nil {
			return fmt.Errorf("%s: %w", action, err)
		}
	}
	return nil
}

func activationGrantSQL(role string) ([]string, error) {
	quoted, err := quotedTenantRole(role)
	if err != nil {
		return nil, err
	}
	return []string{
		"GRANT USAGE ON SCHEMA platform TO " + quoted,
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name) TO " + quoted + " WITH GRANT OPTION",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_step_context_assert_v1(uuid, text, uuid, name) TO " + quoted + " WITH GRANT OPTION",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_step_consume_v1(uuid, text, uuid, name) TO " + quoted + " WITH GRANT OPTION",
	}, nil
}

func activationGrantOptionRevokeSQL(role string) ([]string, error) {
	quoted, err := quotedTenantRole(role)
	if err != nil {
		return nil, err
	}
	return []string{
		"REVOKE GRANT OPTION FOR EXECUTE ON FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name) FROM " + quoted + " CASCADE",
		"REVOKE GRANT OPTION FOR EXECUTE ON FUNCTION platform.tenancy_activation_step_context_assert_v1(uuid, text, uuid, name) FROM " + quoted + " CASCADE",
		"REVOKE GRANT OPTION FOR EXECUTE ON FUNCTION platform.tenancy_activation_step_consume_v1(uuid, text, uuid, name) FROM " + quoted + " CASCADE",
	}, nil
}

func quotedTenantRole(role string) (string, error) {
	if !tenantProvisioningRole.MatchString(role) {
		return "", fmt.Errorf("tenant role %s cannot receive activation routines", role)
	}
	return pgx.Identifier{role}.Sanitize(), nil
}

func adminDatabaseDSN(cfg foundation.Config) (string, error) {
	parsed, err := url.Parse(cfg.AdminDSN)
	if err != nil {
		return "", fmt.Errorf("parse admin dsn: %w", err)
	}
	parsed.Path = "/" + cfg.Database
	return parsed.String(), nil
}
