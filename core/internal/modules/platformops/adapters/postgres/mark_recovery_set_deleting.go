package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) MarkRecoverySetDeleting(
	ctx context.Context,
	recoverySetID string,
	manifestSHA256 string,
	updatedAt time.Time,
) error {
	var changed bool
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		err := database.QueryRow(ctx, markRecoverySetDeletingSQL,
			recoverySetID, manifestSHA256, updatedAt).Scan(&changed)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("mark recovery set deleting: %w", err)
	}
	if !changed {
		return fmt.Errorf("mark recovery set deleting: %w", ports.ErrConflict)
	}
	return nil
}
