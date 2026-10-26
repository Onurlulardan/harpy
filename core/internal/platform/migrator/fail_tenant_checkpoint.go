package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func failTenantCheckpoint(
	ctx context.Context,
	database *sql.DB,
	target postgresstore.TenantMigrationTarget,
	session postgresstore.TenantSessionProof,
	selected postgresv1.SelectedMigration,
	runID, errorCode string,
) error {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	if err := postgresstore.VerifyTenantMigrationSession(
		ctx, transaction, session, target,
	); err != nil {
		return err
	}
	checkpoint, exists, err := postgresstore.LockTenantCheckpoint(
		ctx, transaction, target.TenantID, selected.Owner, selected.Version,
	)
	if err != nil {
		return fmt.Errorf("lock failed tenant checkpoint: %w", err)
	}
	if !exists {
		return fmt.Errorf("failed tenant checkpoint is missing")
	}
	identity := tenantCheckpointIdentity(target, selected, runID, checkpoint.Attempt)
	if !postgresstore.TenantCheckpointMatchesSelection(checkpoint, identity) ||
		checkpoint.RunID != runID || checkpoint.State != "running" {
		return fmt.Errorf("failed tenant checkpoint is not the active attempt")
	}
	if err := postgresstore.CompleteTenantCheckpoint(
		ctx, transaction, identity, "failed", errorCode,
	); err != nil {
		return err
	}
	return transaction.Commit()
}
