package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RejectRestoreExecution(
	ctx context.Context,
	record ports.RestoreFailure,
) (bool, error) {
	return store.change(
		ctx, failSQL, record.RestoreExecutionIdentity, record.Code, record.Fenced,
	)
}
