package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/update_checklist_item.sql
var updateChecklistItemSQL string

func (store *Store) PersistChecklistItem(ctx context.Context, item domain.ChecklistItem, expected int64) error {
	if !item.Valid() || expected < 1 || item.RowVersion != expected+1 {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, updateChecklistItemSQL, item.ID, item.TaskID, item.Text,
		item.Completed, item.Position, item.RowVersion, expected)
}
