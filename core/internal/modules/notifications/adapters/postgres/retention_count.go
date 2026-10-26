package postgres

import (
	"context"
	"fmt"
	"time"
)

const retentionBatchSize = 500

func (store *RetentionStore) count(ctx context.Context, statement, operation string,
	now time.Time,
) (int64, error) {
	var count int64
	err := store.database.QueryRow(
		ctx, statement, now, retentionBatchSize,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	if count < 0 {
		return 0, fmt.Errorf("%s: invalid row count", operation)
	}
	return count, nil
}
