package v1adoption

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
	"github.com/google/uuid"
)

func (store *Store) SealRun(ctx context.Context, runID uuid.UUID,
	checkpoint app.Checkpoint,
) error {
	if !store.usable(ctx) || runID == uuid.Nil || !validCheckpoint(checkpoint) ||
		checkpoint.SourceChecksum != checkpoint.TargetChecksum {
		return app.ErrBlocked
	}
	var completed uuid.UUID
	err := store.queryRow(ctx, completeRunSQL, runID, checkpoint.LastSourceID,
		checkpoint.AdoptedCount, checkpoint.SourceChecksum[:],
		checkpoint.TargetChecksum[:]).Scan(&completed)
	if err != nil || completed != runID {
		return app.ErrBlocked
	}
	return nil
}

var _ app.Store = (*Store)(nil)
