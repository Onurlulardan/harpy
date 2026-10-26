package tasks

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) writeTransition(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID domain.ID, transition domain.Transition,
) (ports.OperationReceipt, error) {
	if !taskID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	target, ok := targetForTransition(transition)
	if !ok {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(taskID.String(), string(transition), ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.task.transition", taskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != meta.ExpectedVersion || task.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			status, err := domain.NextStatus(task.Status, target, transition)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			task.Status, task.RowVersion = status, task.RowVersion+1
			if err = stores.Tasks().PersistTask(ctx, task, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				TaskID, Status string
				Version        int64
			}{
				taskID.String(), string(status), task.RowVersion})
			return ports.OperationReceipt{AggregateID: taskID, RowVersion: task.RowVersion}, payload, nil
		})
}

func targetForTransition(transition domain.Transition) (domain.TaskStatus, bool) {
	switch transition {
	case domain.TransitionComplete:
		return domain.StatusDone, true
	case domain.TransitionCancel:
		return domain.StatusCancelled, true
	case domain.TransitionReopen:
		return domain.StatusOpen, true
	default:
		return "", false
	}
}
