package postgres

import (
	"context"
	"fmt"
	"time"
)

func (store *Store) PruneObservability(
	ctx context.Context,
	environmentID string,
	before time.Time,
	limit int,
) (int64, error) {
	var deleted int64
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		return store.mapDatabaseError(database.QueryRow(
			ctx, pruneObservabilitySQL, environmentID, before, limit,
		).Scan(&deleted))
	})
	if err != nil {
		return 0, fmt.Errorf("prune platform observability: %w", err)
	}
	return deleted, nil
}
