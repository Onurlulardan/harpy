package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_checklist_for_update.sql
var listChecklistForUpdateSQL string

func (store *Store) LoadChecklistForUpdate(ctx context.Context, taskID domain.ID) ([]domain.ChecklistItem, error) {
	if !taskID.Valid() {
		return nil, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return nil, err
	}
	rows, err := store.database.Query(ctx, listChecklistForUpdateSQL, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.ChecklistItem, 0)
	for rows.Next() {
		var id, storedTaskID uuid.UUID
		var text string
		var completed bool
		var position, version int64
		if err = rows.Scan(&id, &storedTaskID, &text, &completed, &position, &version); err != nil {
			return nil, err
		}
		item := domain.ChecklistItem{ID: domain.ID(id), TaskID: domain.ID(storedTaskID), Text: text,
			Completed: completed, Position: position, RowVersion: version}
		if !item.Valid() {
			return nil, application.ErrUnavailable
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil || len(items) > 1000 {
		return nil, application.ErrUnavailable
	}
	return items, nil
}
