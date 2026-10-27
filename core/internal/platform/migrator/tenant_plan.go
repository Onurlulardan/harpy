package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) tenantPlan(
	ctx context.Context,
	connection *sql.Conn,
) ([]postgresv1.SelectedMigration, error) {
	ready, err := postgresstore.TenantInfrastructureExists(ctx, connection)
	if err != nil {
		return nil, err
	}
	if !ready {
		return nil, nil
	}
	targets, err := postgresstore.ListTenantMigrationTargets(ctx, connection)
	if err != nil {
		return nil, err
	}
	pending := make([]postgresv1.SelectedMigration, 0, len(runner.tenantSelected))
	for _, selected := range runner.tenantSelected {
		for _, target := range targets {
			checkpoint, exists, err := postgresstore.ReadTenantCheckpoint(
				ctx, connection, target.TenantID, selected.Owner, selected.Version,
			)
			if err != nil {
				return nil, err
			}
			if !exists {
				pending = append(pending, selected)
				break
			}
			expected := tenantCheckpointIdentity(
				target, selected, checkpoint.RunID, checkpoint.Attempt,
			)
			if !tenantCheckpointMatchesVerifiedSelection(
				checkpoint, expected, selected, runner.tenantAliases,
			) {
				return nil, fmt.Errorf("tenant checkpoint %d differs from catalog", selected.Version)
			}
			if checkpoint.State != "succeeded" {
				pending = append(pending, selected)
				break
			}
		}
	}
	return pending, nil
}
