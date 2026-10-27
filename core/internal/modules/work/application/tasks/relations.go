package tasks

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

const compatibilityLabelColor = "#64748b"

func (service *Service) replaceRelations(ctx context.Context, stores ports.Stores,
	task domain.Task, input RelationInput, assignees []domain.Principal, revision int64,
) error {
	if input.ReplaceAssignees {
		allowed, err := stores.Boards().AssigneesHaveAccess(ctx, task.BoardID, assignees)
		if err != nil || !allowed {
			return ports.ErrForbidden
		}
		if err = stores.Tasks().ReplaceAssignees(ctx, task.ID, assignees, revision); err != nil {
			return err
		}
	}
	if input.ReplaceLabels {
		if err := service.replaceLabels(ctx, stores.Tasks(), task, input.Labels); err != nil {
			return err
		}
	}
	if input.ReplaceChecklist {
		return service.replaceChecklist(ctx, stores.Tasks(), task.ID, input.Checklist)
	}
	return nil
}

func (service *Service) replaceLabels(ctx context.Context, store ports.TaskStore,
	task domain.Task, names []string,
) error {
	labels, err := store.LoadLabelsForUpdate(ctx, task.BoardID)
	if err != nil {
		return err
	}
	byName := make(map[string]domain.ID, len(labels))
	for _, label := range labels {
		byName[label.Name] = label.ID
	}
	desired := make(map[domain.ID]bool, len(names))
	for _, name := range names {
		id, exists := byName[name]
		if !exists {
			if len(labels) >= 100 {
				return domain.ErrBoardLimit
			}
			id, err = service.ids.New(service.clock.Now().UTC())
			if err != nil {
				return ports.ErrUnavailable
			}
			label := domain.Label{ID: id, BoardID: task.BoardID, Name: name,
				Color: compatibilityLabelColor, RowVersion: 1}
			if err = store.InsertLabel(ctx, label); err != nil {
				return err
			}
			labels = append(labels, label)
		}
		desired[id] = true
	}
	current, err := store.LoadTaskLabelIDsForUpdate(ctx, task.ID)
	if err != nil {
		return err
	}
	for _, id := range current {
		if !desired[id] {
			if err = store.PersistTaskLabelLink(ctx, task.ID, id, false); err != nil {
				return err
			}
		}
		delete(desired, id)
	}
	for id := range desired {
		if err = store.PersistTaskLabelLink(ctx, task.ID, id, true); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) replaceChecklist(ctx context.Context, store ports.TaskStore,
	taskID domain.ID, values []ChecklistValue,
) error {
	existing, err := store.LoadChecklistForUpdate(ctx, taskID)
	if err != nil {
		return err
	}
	byID, keep, position := map[domain.ID]domain.ChecklistItem{}, map[domain.ID]bool{}, int64(0)
	for _, item := range existing {
		byID[item.ID] = item
		if item.Position > position {
			position = item.Position
		}
	}
	for _, value := range values {
		if value.ID == nil {
			position += domain.SparseGap
			id, idErr := service.ids.New(service.clock.Now().UTC())
			if idErr != nil {
				return ports.ErrUnavailable
			}
			item := domain.ChecklistItem{ID: id, TaskID: taskID, Text: value.Text,
				Completed: value.Completed, Position: position, RowVersion: 1}
			if err = store.InsertChecklistItem(ctx, item); err != nil {
				return err
			}
			continue
		}
		item, found := byID[*value.ID]
		if !found {
			return ports.ErrInvalid
		}
		keep[item.ID] = true
		if item.Text != value.Text || item.Completed != value.Completed {
			old := item.RowVersion
			item.Text, item.Completed, item.RowVersion = value.Text, value.Completed, old+1
			if err = store.PersistChecklistItem(ctx, item, old); err != nil {
				return err
			}
		}
	}
	now := service.clock.Now().UTC()
	for _, item := range existing {
		if !keep[item.ID] {
			if err = store.PersistChecklistArchive(ctx, item.ID, item.RowVersion,
				timePointer(now)); err != nil {
				return err
			}
		}
	}
	return nil
}

func timePointer(value time.Time) *time.Time { return &value }
