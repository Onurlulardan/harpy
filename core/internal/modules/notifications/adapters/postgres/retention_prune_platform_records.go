package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (store *RetentionStore) prunePlatformRecords(ctx context.Context, tenantID uuid.UUID,
	now time.Time,
) (int64, error) {
	var count int64
	err := store.database.QueryRow(ctx, purgePlatformRecordsSQL,
		tenantID, now, retentionBatchSize).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("purge notification platform records: %w", err)
	}
	if count < 0 {
		return 0, fmt.Errorf("purge notification platform records: invalid row count")
	}
	return count, nil
}
