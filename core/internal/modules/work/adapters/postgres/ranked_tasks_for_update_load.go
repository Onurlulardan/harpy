package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_ranked_tasks_for_update.sql
var listRankedTasksForUpdateSQL string

func (store *Store) LoadRankedTasksForUpdate(ctx context.Context, columnID domain.ID) ([]domain.RankedTask, error) {
	if !columnID.Valid() {
		return nil, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return nil, err
	}
	rows, err := store.database.Query(ctx, listRankedTasksForUpdateSQL, columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.RankedTask, 0)
	for rows.Next() {
		var id uuid.UUID
		var position int64
		if err = rows.Scan(&id, &position); err != nil {
			return nil, err
		}
		items = append(items, domain.RankedTask{ID: domain.ID(id), Position: position})
	}
	if err = rows.Err(); err != nil || len(items) > domain.MaxColumnTasks {
		return nil, application.ErrUnavailable
	}
	return items, nil
}
