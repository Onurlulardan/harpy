package tasks

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type MoveTaskInput struct {
	TaskID, ColumnID   domain.ID
	Before, After      *domain.ID
	CompatibilityOrder *int
}

func (service *Service) MoveTask(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input MoveTaskInput,
) (ports.OperationReceipt, error) {
	if !input.TaskID.Valid() || !input.ColumnID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	if input.CompatibilityOrder != nil && (input.Before != nil || input.After != nil) {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	var hash [32]byte
	plan := func(ranked []domain.RankedTask) (domain.PositionPlan, error) {
		return domain.PlanPosition(ranked, input.TaskID,
			domain.MoveNeighbors{Before: input.Before, After: input.After})
	}
	if input.CompatibilityOrder == nil {
		hash = ports.Fingerprint(input.TaskID.String(), input.ColumnID.String(),
			ports.IDString(input.Before), ports.IDString(input.After),
			ports.VersionString(meta.ExpectedVersion))
	} else {
		order := *input.CompatibilityOrder
		hash = ports.Fingerprint("work-v1-order.v1", input.TaskID.String(),
			input.ColumnID.String(), ports.VersionString(int64(order)),
			ports.VersionString(meta.ExpectedVersion))
		plan = func(ranked []domain.RankedTask) (domain.PositionPlan, error) {
			return domain.PlanPosition(ranked, input.TaskID,
				legacyOrderNeighbors(ranked, input.TaskID, order))
		}
	}
	return service.moveTask(ctx, scope, meta, input.TaskID, input.ColumnID, hash,
		plan)
}

func legacyOrderNeighbors(ranked []domain.RankedTask, moving domain.ID,
	order int,
) domain.MoveNeighbors {
	remaining := make([]domain.RankedTask, 0, len(ranked))
	for _, task := range ranked {
		if task.ID != moving {
			remaining = append(remaining, task)
		}
	}
	if len(remaining) == 0 {
		return domain.MoveNeighbors{}
	}
	if order <= 0 {
		return domain.MoveNeighbors{After: &remaining[0].ID}
	}
	if order >= len(remaining) {
		return domain.MoveNeighbors{Before: &remaining[len(remaining)-1].ID}
	}
	return domain.MoveNeighbors{Before: &remaining[order-1].ID,
		After: &remaining[order].ID}
}

func (service *Service) moveTask(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID, columnID domain.ID, hash [32]byte,
	planPosition func([]domain.RankedTask) (domain.PositionPlan, error),
) (ports.OperationReceipt, error) {
	return service.mutator.Mutate(ctx, scope, "work.task.move", taskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != meta.ExpectedVersion || task.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			column, err := loadMoveColumns(ctx, stores.Boards(), task.ColumnID, columnID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if column.BoardID != task.BoardID {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			ranked, err := stores.Tasks().LoadRankedTasksForUpdate(ctx, columnID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			plan, err := planPosition(ranked)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			status, err := domain.NextStatus(task.Status, column.Status, domain.TransitionMove)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if len(plan.Rebalanced) > 0 {
				if err = stores.Tasks().RebalanceTasks(ctx, columnID, plan.Rebalanced); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
			}
			task.ColumnID, task.Position, task.Status = columnID, plan.Position, status
			task.RowVersion++
			if err = stores.Tasks().PersistTask(ctx, task, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				TaskID, ColumnID  string
				Position, Version int64
			}{
				taskID.String(), columnID.String(), plan.Position, task.RowVersion})
			return ports.OperationReceipt{AggregateID: taskID, RowVersion: task.RowVersion}, payload, nil
		})
}

func loadMoveColumns(ctx context.Context, store ports.BoardStore,
	sourceID, targetID domain.ID) (domain.Column, error) {
	if sourceID == targetID {
		return store.LoadColumnForUpdate(ctx, targetID)
	}
	first, second := sourceID, targetID
	if domain.CompareID(first, second) > 0 {
		first, second = second, first
	}
	firstColumn, err := store.LoadColumnForUpdate(ctx, first)
	if err != nil {
		return domain.Column{}, err
	}
	secondColumn, err := store.LoadColumnForUpdate(ctx, second)
	if err != nil {
		return domain.Column{}, err
	}
	if firstColumn.ID == targetID {
		return firstColumn, nil
	}
	return secondColumn, nil
}
