package labels

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) SetTaskLabel(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, taskID, labelID domain.ID, linked bool,
) (ports.OperationReceipt, error) {
	if !taskID.Valid() || !labelID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(taskID.String(), labelID.String(), ports.VersionString(meta.ExpectedVersion),
		ports.VersionString(boolVersion(linked)))
	return service.mutator.Mutate(ctx, scope, "work.label.change", taskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != meta.ExpectedVersion {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			label, err := stores.Tasks().LoadLabelForUpdate(ctx, labelID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if label.BoardID != task.BoardID || label.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Tasks().PersistTaskLabelLink(ctx, taskID, labelID, linked); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			task.RowVersion++
			if err = stores.Tasks().PersistTask(ctx, task, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			return ports.OperationReceipt{AggregateID: taskID, RowVersion: task.RowVersion}, []byte(`{"changed":true}`), nil
		})
}

func boolVersion(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
