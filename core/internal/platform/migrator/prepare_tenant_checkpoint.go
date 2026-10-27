package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func prepareTenantCheckpoint(
	ctx context.Context,
	database *sql.DB,
	target postgresstore.TenantMigrationTarget,
	session postgresstore.TenantSessionProof,
	selected postgresv1.SelectedMigration,
	aliases map[int64]historySupersession,
	runID string,
) (bool, error) {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tenant checkpoint: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()
	if err := postgresstore.VerifyTenantMigrationSession(
		ctx, transaction, session, target,
	); err != nil {
		return false, err
	}
	checkpoint, exists, err := postgresstore.LockTenantCheckpoint(
		ctx, transaction, target.TenantID, selected.Owner, selected.Version,
	)
	if err != nil {
		return false, err
	}
	if !exists {
		identity := tenantCheckpointIdentity(target, selected, runID, 1)
		if err := postgresstore.InsertTenantCheckpoint(ctx, transaction, identity); err != nil {
			return false, err
		}
		checkpoint = postgresstore.TenantCheckpoint{
			TenantCheckpointIdentity: identity, State: "pending",
		}
	}
	expected := tenantCheckpointIdentity(
		target, selected, checkpoint.RunID, checkpoint.Attempt,
	)
	if !tenantCheckpointMatchesVerifiedSelection(
		checkpoint, expected, selected, aliases,
	) {
		return false, fmt.Errorf("tenant checkpoint differs from verified selection")
	}
	if checkpoint.State == "succeeded" {
		return false, transaction.Commit()
	}
	checkpoint, err = reconcileTenantCheckpoint(
		ctx, transaction, checkpoint, runID,
	)
	if err != nil {
		return false, err
	}
	storedCheckpoint := checkpoint
	if storedCheckpoint.State == "failed" {
		storedCheckpoint.Attempt--
	}
	identity := tenantCheckpointIdentity(target, selected, runID, checkpoint.Attempt)
	if err := postgresstore.StartTenantCheckpoint(
		ctx, transaction, identity, storedCheckpoint,
	); err != nil {
		return false, err
	}
	if err := transaction.Commit(); err != nil {
		return false, fmt.Errorf("commit tenant checkpoint start: %w", err)
	}
	return true, nil
}
