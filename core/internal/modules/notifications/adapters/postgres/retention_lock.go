package postgres

import (
	"context"
	"fmt"
)

func (store *RetentionStore) lock(ctx context.Context) (int64, error) {
	var revision int64
	err := store.database.QueryRow(ctx, lockCheckpointSQL).Scan(&revision)
	if err != nil {
		return 0, fmt.Errorf("lock notification retention checkpoint: %w", err)
	}
	if revision < 0 {
		return 0, fmt.Errorf("lock notification retention checkpoint: invalid revision")
	}
	return revision, nil
}
