package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) ExecuteApply(
	ctx context.Context,
	record ports.RestoreApplyEvidence,
) error {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || !record.RestoreExecutionIdentity.Valid() ||
		!digestPattern.MatchString(record.DrainAckSHA256) ||
		!digestPattern.MatchString(record.ControlDatabaseLocatorSHA256) ||
		!digestPattern.MatchString(record.ScratchDatabaseLocatorSHA256) ||
		record.ControlDatabaseLocatorSHA256 ==
			record.ScratchDatabaseLocatorSHA256 {
		return ErrTransition
	}
	value, err := service.store.ExecuteRestoreApply(ctx, record)
	return changed(err, value)
}
