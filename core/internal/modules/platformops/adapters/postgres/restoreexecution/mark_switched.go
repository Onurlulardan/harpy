package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) MarkRestoreSwitched(
	ctx context.Context,
	record ports.RestoreSwitchEvidence,
) (bool, error) {
	return store.change(
		ctx, markSwitchedSQL, record.RestoreExecutionIdentity,
		record.ObjectMerkleSHA256, record.LocatorRemapSHA256,
		record.SafetyEvidenceSHA256, record.SwitchAckSHA256,
		record.PreviousDatabaseSystemID, record.NewDatabaseSystemID,
		record.ControlDatabaseLocatorSHA256,
		record.ScratchDatabaseLocatorSHA256,
	)
}
