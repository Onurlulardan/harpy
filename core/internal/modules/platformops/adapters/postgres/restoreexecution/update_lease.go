package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) UpdateRestoreLease(
	ctx context.Context,
	identity ports.RestoreExecutionIdentity,
) (bool, error) {
	return store.change(ctx, heartbeatSQL, identity)
}
