package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/insert_checklist_item.sql
var insertChecklistItemSQL string

func (store *Store) InsertChecklistItem(ctx context.Context, item domain.ChecklistItem) error {
	if !item.Valid() {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, insertChecklistItemSQL, item.ID, item.TaskID, item.Text,
		item.Completed, item.Position, item.RowVersion)
}
