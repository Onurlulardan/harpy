package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

type tenantCheckpointReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func verifyTenantCheckpoint(
	ctx context.Context,
	database tenantCheckpointReader,
	target postgresstore.TenantMigrationTarget,
	selected postgresv1.SelectedMigration,
	aliases map[int64]historySupersession,
) error {
	checkpoint, exists, err := postgresstore.ReadTenantCheckpoint(
		ctx, database, target.TenantID, selected.Owner, selected.Version,
	)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("tenant checkpoint %d is missing", selected.Version)
	}
	expected := tenantCheckpointIdentity(
		target, selected, checkpoint.RunID, checkpoint.Attempt,
	)
	if !tenantCheckpointMatchesVerifiedSelection(
		checkpoint, expected, selected, aliases,
	) ||
		checkpoint.State != "succeeded" {
		return fmt.Errorf("tenant checkpoint %d is not exact", selected.Version)
	}
	return nil
}
