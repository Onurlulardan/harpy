package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) ValidateRestoreExecution(
	ctx context.Context,
	record ports.RestoreVerification,
) (bool, error) {
	return store.change(
		ctx, beginVerifySQL, record.RestoreExecutionIdentity,
		record.ArchiveReadbackSHA256, record.NewDatabaseSystemID,
	)
}
