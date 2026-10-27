package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) ValidateRestore(
	ctx context.Context,
	record ports.RestoreVerification,
) error {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || !record.RestoreExecutionIdentity.Valid() ||
		!digestPattern.MatchString(record.ArchiveReadbackSHA256) ||
		!canonicalUUID.MatchString(record.NewDatabaseSystemID) {
		return ErrTransition
	}
	value, err := service.store.ValidateRestoreExecution(ctx, record)
	return changed(err, value)
}
