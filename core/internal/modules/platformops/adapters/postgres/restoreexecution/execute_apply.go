package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) ExecuteRestoreApply(
	ctx context.Context,
	record ports.RestoreApplyEvidence,
) (bool, error) {
	return store.change(
		ctx, beginApplySQL, record.RestoreExecutionIdentity,
		record.DrainAckSHA256, record.ControlDatabaseLocatorSHA256,
		record.ScratchDatabaseLocatorSHA256,
	)
}
