package v1adoption

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
	"github.com/google/uuid"
)

func (store *Store) SaveCheckpoint(ctx context.Context, runID uuid.UUID,
	checkpoint app.Checkpoint,
) error {
	if !store.usable(ctx) || runID == uuid.Nil || !validCheckpoint(checkpoint) {
		return app.ErrInvalid
	}
	if err := store.execute(ctx, saveCheckpointSQL, runID,
		checkpoint.LastSourceID, checkpoint.AdoptedCount,
		checkpoint.SourceChecksum[:], checkpoint.TargetChecksum[:]); err != nil {
		return app.ErrUnavailable
	}
	return nil
}

func validCheckpoint(value app.Checkpoint) bool {
	return value.AdoptedCount >= 0 && value.SourceChecksum != [32]byte{} &&
		value.TargetChecksum != [32]byte{}
}
