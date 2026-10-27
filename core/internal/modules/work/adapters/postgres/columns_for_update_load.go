package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_columns_for_update.sql
var listColumnsForUpdateSQL string

func (store *Store) LoadColumnsForUpdate(ctx context.Context, boardID domain.ID) ([]domain.Column, error) {
	if !boardID.Valid() {
		return nil, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return nil, err
	}
	rows, err := store.database.Query(ctx, listColumnsForUpdateSQL, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := make([]domain.Column, 0)
	for rows.Next() {
		var id, storedBoardID uuid.UUID
		var name, color, status string
		var position, version int64
		if err = rows.Scan(&id, &storedBoardID, &name, &color,
			&status, &position, &version); err != nil {
			return nil, err
		}
		column := domain.Column{ID: domain.ID(id), BoardID: domain.ID(storedBoardID),
			Name: name, Color: color, Status: domain.TaskStatus(status),
			Position: position, RowVersion: version}
		if !column.Valid() {
			return nil, application.ErrUnavailable
		}
		columns = append(columns, column)
	}
	if err = rows.Err(); err != nil || len(columns) > domain.MaxBoardColumns {
		return nil, application.ErrUnavailable
	}
	return columns, nil
}
