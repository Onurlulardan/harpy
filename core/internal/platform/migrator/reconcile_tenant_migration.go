package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func reconcileTenantMigration(
	ctx context.Context,
	database *sql.DB,
	target postgresstore.TenantMigrationTarget,
	session postgresstore.TenantSessionProof,
	selected postgresv1.SelectedMigration,
	aliases map[int64]historySupersession,
) (committed bool, safelyFailed bool, err error) {
	transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return false, false, err
	}
	defer transaction.Rollback()
	if err := postgresstore.VerifyTenantMigrationSession(
		ctx, transaction, session, target,
	); err != nil {
		return false, false, err
	}
	checkpoint, exists, err := postgresstore.ReadTenantCheckpoint(
		ctx, transaction, target.TenantID, selected.Owner, selected.Version,
	)
	if err != nil {
		return false, false, err
	}
	if !exists {
		return false, false, fmt.Errorf("tenant checkpoint disappeared during reconciliation")
	}
	expected := tenantCheckpointIdentity(
		target, selected, checkpoint.RunID, checkpoint.Attempt,
	)
	if !tenantCheckpointMatchesVerifiedSelection(
		checkpoint, expected, selected, aliases,
	) {
		return false, false, fmt.Errorf("tenant checkpoint changed during reconciliation")
	}
	if err := transaction.Commit(); err != nil {
		return false, false, err
	}
	switch checkpoint.State {
	case "succeeded":
		return true, false, nil
	case "running":
		return false, true, nil
	case "failed":
		return false, false, nil
	default:
		return false, false, fmt.Errorf("tenant checkpoint reconciliation state is %s", checkpoint.State)
	}
}
