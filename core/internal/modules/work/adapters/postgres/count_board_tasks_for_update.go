package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/count_board_tasks.sql
var countBoardTasksSQL string

func (store *Store) CountBoardTasksForUpdate(ctx context.Context, boardID domain.ID) (int, error) {
	if !boardID.Valid() {
		return 0, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return 0, err
	}
	var count int
	if err := store.database.QueryRow(ctx, countBoardTasksSQL, boardID).Scan(&count); err != nil {
		return 0, err
	}
	if count < 0 || count > domain.MaxBoardTasks {
		return 0, application.ErrUnavailable
	}
	return count, nil
}
