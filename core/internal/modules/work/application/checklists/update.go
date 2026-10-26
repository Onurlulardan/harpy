package checklists

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type UpdateChecklistInput struct {
	Item        domain.ChecklistItem
	TaskVersion int64
}

func (service *Service) UpdateChecklistItem(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input UpdateChecklistInput,
) (ports.OperationReceipt, error) {
	input.Item.Text = strings.TrimSpace(input.Item.Text)
	if !input.Item.ID.Valid() || !input.Item.TaskID.Valid() || input.TaskVersion < 1 ||
		meta.ExpectedVersion < 1 || input.Item.DeletedAt != nil {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	input.Item.RowVersion = meta.ExpectedVersion + 1
	if !input.Item.Valid() {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(input.Item.ID.String(), input.Item.Text,
		ports.VersionString(boolVersion(input.Item.Completed)),
		ports.VersionString(int64(input.Item.Position)),
		ports.VersionString(meta.ExpectedVersion), ports.VersionString(input.TaskVersion))
	return service.mutator.Mutate(ctx, scope, "work.checklist.change", input.Item.TaskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, input.Item.TaskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != input.TaskVersion {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			if err = stores.Tasks().PersistChecklistItem(ctx, input.Item, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			task.RowVersion++
			if err = stores.Tasks().PersistTask(ctx, task, input.TaskVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				TaskID, ItemID string
				Completed      bool
			}{
				input.Item.TaskID.String(), input.Item.ID.String(), input.Item.Completed})
			return ports.OperationReceipt{AggregateID: input.Item.ID, RowVersion: input.Item.RowVersion}, payload, nil
		})
}

func boolVersion(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
