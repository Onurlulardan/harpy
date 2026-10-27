package postgres

import (
	"context"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *Store) Insert(
	ctx context.Context,
	db goosedb.DBTxConn,
	request goosedb.InsertRequest,
) error {
	if request.Version == 0 {
		if _, err := db.ExecContext(ctx, insertAnchorSQL); err != nil {
			return fmt.Errorf("insert migration history anchor: %w", err)
		}
		return nil
	}
	metadata, exists := store.metadataByVersion[request.Version]
	if !exists || request.Version < 1 {
		return fmt.Errorf("migration %d has no validated history metadata", request.Version)
	}
	if _, err := db.ExecContext(
		ctx, insertMigrationSQL,
		metadata.version, metadata.owner, metadata.scope, metadata.origin,
		metadata.selectedPath, metadata.selectedSHA256, metadata.runID,
		metadata.targetDatabaseGeneration, metadata.nextMinimumRuntimeGeneration,
	); err != nil {
		return fmt.Errorf("insert migration history version %d: %w", request.Version, err)
	}
	return nil
}
