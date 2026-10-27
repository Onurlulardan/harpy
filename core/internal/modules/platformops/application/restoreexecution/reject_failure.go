package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) RejectFailure(
	ctx context.Context,
	record ports.RestoreFailure,
) error {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || !record.RestoreExecutionIdentity.Valid() ||
		!codePattern.MatchString(record.Code) {
		return ErrTransition
	}
	value, err := service.store.RejectRestoreExecution(ctx, record)
	return changed(err, value)
}
