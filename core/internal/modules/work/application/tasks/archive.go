package tasks

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) ArchiveTask(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID domain.ID) (ports.OperationReceipt, error) {
	return service.writeArchived(ctx, scope, meta, taskID, true, "work.task.archive")
}

func (service *Service) writeArchived(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID domain.ID, archived bool, authority string,
) (ports.OperationReceipt, error) {
	if !taskID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(taskID.String(), authority, ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, authority, taskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != meta.ExpectedVersion || (task.DeletedAt != nil) == archived {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			if archived {
				now := service.clock.Now().UTC()
				task.DeletedAt = &now
			} else {
				task.DeletedAt = nil
			}
			task.RowVersion++
			if err = stores.Tasks().PersistTask(ctx, task, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				TaskID   string
				Archived bool
				Version  int64
			}{
				taskID.String(), archived, task.RowVersion})
			return ports.OperationReceipt{AggregateID: taskID, RowVersion: task.RowVersion}, payload, nil
		})
}
