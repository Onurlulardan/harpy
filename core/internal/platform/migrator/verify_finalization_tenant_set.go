package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) verifyFinalizationTenantSet(
	ctx context.Context,
	transaction *sql.Tx,
) error {
	if err := postgresstore.LockTenantMigrationTargets(ctx, transaction); err != nil {
		return err
	}
	targets, err := postgresstore.ListTenantMigrationTargets(ctx, transaction)
	if err != nil {
		return err
	}
	if len(targets) != 0 && len(runner.tenantSelected) == 0 {
		return fmt.Errorf("tenant migration selection is absent during finalization")
	}
	for _, target := range targets {
		for _, selected := range runner.tenantSelected {
			if err := verifyTenantCheckpoint(
				ctx, transaction, target, selected, runner.tenantAliases,
			); err != nil {
				return fmt.Errorf("verify finalization tenant %s: %w", target.TenantID, err)
			}
		}
	}
	return nil
}
