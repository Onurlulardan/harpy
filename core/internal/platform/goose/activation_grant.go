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

func grantActivationRoutines(
	ctx context.Context,
	cfg foundation.Config,
	role string,
) error {
	statements, err := activationGrantSQL(role)
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
	for _, statement := range statements {
		if _, err = connection.Exec(ctx, statement); err != nil {
			return fmt.Errorf("grant activation routine to tenant role: %w", err)
		}
	}
	return nil
}

func activationGrantSQL(role string) ([]string, error) {
	if !tenantProvisioningRole.MatchString(role) {
		return nil, fmt.Errorf("tenant role %s cannot receive activation routines", role)
	}
	quoted := pgx.Identifier{role}.Sanitize()
	return []string{
		"GRANT USAGE ON SCHEMA platform TO " + quoted,
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name) TO " + quoted,
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_step_context_assert_v1(uuid, text, uuid, name) TO " + quoted,
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_step_consume_v1(uuid, text, uuid, name) TO " + quoted,
	}, nil
}

func adminDatabaseDSN(cfg foundation.Config) (string, error) {
	parsed, err := url.Parse(cfg.AdminDSN)
	if err != nil {
		return "", fmt.Errorf("parse admin dsn: %w", err)
	}
	parsed.Path = "/" + cfg.Database
	return parsed.String(), nil
}
