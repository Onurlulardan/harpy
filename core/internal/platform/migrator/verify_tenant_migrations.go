package migrator

import (
	"context"
	"database/sql"
	"fmt"
)

func (runner *Runner) verifyTenantMigrations(
	ctx context.Context,
	connection *sql.Conn,
) error {
	if len(runner.tenantSelected) == 0 {
		return nil
	}
	ready, err := tenantInfrastructureRequired(ctx, connection)
	if err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("tenant migration infrastructure is absent")
	}
	targets, err := listTenantTargets(ctx, connection)
	if err != nil {
		return err
	}
	for _, target := range targets {
		if err := runner.verifyTenantTarget(ctx, target); err != nil {
			return fmt.Errorf("verify tenant %s: %w", target.TenantID, err)
		}
	}
	return nil
}
