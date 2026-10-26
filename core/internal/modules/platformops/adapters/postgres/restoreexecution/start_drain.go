package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) StartRestoreDrain(
	ctx context.Context,
	identity ports.RestoreExecutionIdentity,
) (bool, error) {
	return store.change(ctx, beginDrainSQL, identity)
}
