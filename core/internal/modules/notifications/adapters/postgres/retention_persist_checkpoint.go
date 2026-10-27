package postgres

import (
	"context"
	"fmt"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
)

func (store *RetentionStore) persistCheckpoint(ctx context.Context, cycle retentionapp.Cycle,
	revision int64,
) error {
	var stored int64
	err := store.database.QueryRow(ctx, updateCheckpointSQL, revision+1,
		cycle.Claim.JobID, cycle.Claim.AttemptID, cycle.CompletedAt, revision,
	).Scan(&stored)
	if err != nil {
		return fmt.Errorf("update notification retention checkpoint: %w", err)
	}
	if stored != revision+1 {
		return fmt.Errorf("update notification retention checkpoint: revision mismatch")
	}
	return nil
}
