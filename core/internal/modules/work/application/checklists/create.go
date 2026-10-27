package checklists

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type CreateChecklistInput struct {
	TaskID        domain.ID
	TaskVersion   int64
	Text          string
	Before, After *domain.ID
}

func (service *Service) CreateChecklistItem(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input CreateChecklistInput,
) (ports.OperationReceipt, error) {
	input.Text = strings.TrimSpace(input.Text)
	if !input.TaskID.Valid() || input.TaskVersion < 1 || meta.ExpectedVersion != 0 ||
		input.Text == "" || len(input.Text) > 1000 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(input.TaskID.String(), input.Text, ports.IDString(input.Before),
		ports.IDString(input.After), ports.VersionString(input.TaskVersion))
	return service.mutator.Mutate(ctx, scope, "work.checklist.change", input.TaskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, input.TaskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != input.TaskVersion {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			items, err := stores.Tasks().LoadChecklistForUpdate(ctx, input.TaskID)
			if err != nil || len(items) >= 1000 {
				return ports.OperationReceipt{}, nil, domain.ErrBoardLimit
			}
			ranked := make([]domain.RankedTask, 0, len(items))
			for _, item := range items {
				ranked = append(ranked, domain.RankedTask{ID: item.ID, Position: item.Position})
			}
			id, err := service.ids.New(service.clock.Now().UTC())
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			plan, err := domain.PlanPosition(ranked, id, domain.MoveNeighbors{Before: input.Before, After: input.After})
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			for _, rankedItem := range plan.Rebalanced {
				for _, item := range items {
					if item.ID == rankedItem.ID {
						old := item.RowVersion
						item.Position, item.RowVersion = rankedItem.Position, old+1
						if err = stores.Tasks().PersistChecklistItem(ctx, item, old); err != nil {
							return ports.OperationReceipt{}, nil, err
						}
					}
				}
			}
			item := domain.ChecklistItem{ID: id, TaskID: input.TaskID, Text: input.Text,
				Position: plan.Position, RowVersion: 1}
			if !item.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Tasks().InsertChecklistItem(ctx, item); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			task.RowVersion++
			if err = stores.Tasks().PersistTask(ctx, task, input.TaskVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct{ TaskID, ItemID string }{input.TaskID.String(), id.String()})
			return ports.OperationReceipt{AggregateID: id, RowVersion: 1}, payload, nil
		})
}
