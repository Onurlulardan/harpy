package postgres

import (
	"context"
	"database/sql"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *TenantStore) Insert(
	ctx context.Context,
	database goosedb.DBTxConn,
	request goosedb.InsertRequest,
) error {
	transaction, ok := database.(*sql.Tx)
	if !ok || transaction == nil {
		return fmt.Errorf("tenant migration receipt requires an active transaction")
	}
	metadata, exists := store.metadataByVersion[request.Version]
	if !exists || request.Version < 1 {
		return fmt.Errorf("tenant migration %d has no validated metadata", request.Version)
	}
	if err := VerifyTenantMigrationSession(
		ctx, transaction, store.session, store.target,
	); err != nil {
		return err
	}
	checkpoint, found, err := LockTenantCheckpoint(
		ctx, transaction, store.target.TenantID, metadata.owner, metadata.version,
	)
	if err != nil {
		return fmt.Errorf("lock running tenant checkpoint: %w", err)
	}
	if !found {
		return fmt.Errorf("running tenant checkpoint is missing")
	}
	identity := metadata.identity(store.target, checkpoint.Attempt)
	if !tenantCheckpointMatches(checkpoint, identity) || checkpoint.State != "running" {
		return fmt.Errorf("running tenant checkpoint differs from selected migration")
	}
	if request.Version == store.finalVersion {
		if err := InsertTenantSchemaIdentity(
			ctx, transaction, store.target, metadata.sentinelOrigin(),
		); err != nil {
			return fmt.Errorf("write tenant schema identity: %w", err)
		}
		if err := VerifyTenantSchemaIdentity(
			ctx, transaction, store.target, metadata.sentinelOrigin(),
		); err != nil {
			return fmt.Errorf("verify tenant schema identity: %w", err)
		}
	}
	return CompleteTenantCheckpoint(ctx, transaction, identity, "succeeded", nil)
}
