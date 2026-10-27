package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/update_task.sql
var updateTaskSQL string

func (store *Store) PersistTask(ctx context.Context, task domain.Task, expected int64) error {
	if !task.Valid() || expected < 1 || task.RowVersion != expected+1 {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, updateTaskSQL, task.ID, task.ColumnID, task.Status,
		task.Priority, task.Title, task.Description, task.Position, task.RowVersion,
		task.StartsAt, task.DueAt, task.DeletedAt, expected)
}
