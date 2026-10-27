package checklists

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) ArchiveChecklistItem(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID, itemID domain.ID, taskVersion int64,
) (ports.OperationReceipt, error) {
	return service.writeArchived(ctx, scope, meta, taskID, itemID, taskVersion, true)
}

func (service *Service) writeArchived(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID, itemID domain.ID, taskVersion int64, archived bool,
) (ports.OperationReceipt, error) {
	if !taskID.Valid() || !itemID.Valid() || meta.ExpectedVersion < 1 || taskVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(taskID.String(), itemID.String(), ports.VersionString(meta.ExpectedVersion),
		ports.VersionString(taskVersion), ports.VersionString(boolVersion(archived)))
	return service.mutator.Mutate(ctx, scope, "work.checklist.change", taskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != taskVersion {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			var deletedAt *time.Time
			if archived {
				now := service.clock.Now().UTC()
				deletedAt = &now
			}
			if err = stores.Tasks().PersistChecklistArchive(ctx, itemID,
				meta.ExpectedVersion, deletedAt); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			task.RowVersion++
			if err = stores.Tasks().PersistTask(ctx, task, taskVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			return ports.OperationReceipt{AggregateID: itemID,
				RowVersion: meta.ExpectedVersion + 1}, []byte(`{"archived":true}`), nil
		})
}
