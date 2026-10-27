package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_task_label_ids_for_update.sql
var listTaskLabelIDsForUpdateSQL string

func (store *Store) LoadTaskLabelIDsForUpdate(ctx context.Context,
	taskID domain.ID,
) ([]domain.ID, error) {
	if !taskID.Valid() {
		return nil, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return nil, err
	}
	rows, err := store.database.Query(ctx, listTaskLabelIDsForUpdateSQL, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.ID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, domain.ID(id))
	}
	return result, rows.Err()
}
