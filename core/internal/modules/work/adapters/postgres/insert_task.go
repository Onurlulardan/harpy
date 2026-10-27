package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/insert_task.sql
var insertTaskSQL string

func (store *Store) InsertTask(ctx context.Context, task domain.Task) error {
	if !task.Valid() || task.DeletedAt != nil {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, insertTaskSQL, task.ID, task.BoardID, task.ColumnID,
		task.CreatorID, task.Kind, task.Status, task.Priority, task.Title, task.Description,
		task.Position, task.RowVersion, task.StartsAt, task.DueAt)
}
