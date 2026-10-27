package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) MarkSwitched(
	ctx context.Context,
	record ports.RestoreSwitchEvidence,
) error {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || !record.RestoreExecutionIdentity.Valid() ||
		!digestPattern.MatchString(record.ObjectMerkleSHA256) ||
		!digestPattern.MatchString(record.LocatorRemapSHA256) ||
		!digestPattern.MatchString(record.SafetyEvidenceSHA256) ||
		!digestPattern.MatchString(record.SwitchAckSHA256) ||
		!canonicalUUID.MatchString(record.PreviousDatabaseSystemID) ||
		!canonicalUUID.MatchString(record.NewDatabaseSystemID) ||
		record.PreviousDatabaseSystemID == record.NewDatabaseSystemID ||
		!digestPattern.MatchString(record.ControlDatabaseLocatorSHA256) ||
		!digestPattern.MatchString(record.ScratchDatabaseLocatorSHA256) ||
		record.ControlDatabaseLocatorSHA256 ==
			record.ScratchDatabaseLocatorSHA256 {
		return ErrTransition
	}
	value, err := service.store.MarkRestoreSwitched(ctx, record)
	return changed(err, value)
}
