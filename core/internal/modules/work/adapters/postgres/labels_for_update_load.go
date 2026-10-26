package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_labels_for_update.sql
var listLabelsForUpdateSQL string

func (store *Store) LoadLabelsForUpdate(ctx context.Context, boardID domain.ID) ([]domain.Label, error) {
	if !boardID.Valid() {
		return nil, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return nil, err
	}
	rows, err := store.database.Query(ctx, listLabelsForUpdateSQL, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.Label, 0)
	for rows.Next() {
		var id, storedBoardID uuid.UUID
		var name, color string
		var version int64
		if err = rows.Scan(&id, &storedBoardID, &name, &color, &version); err != nil {
			return nil, err
		}
		label := domain.Label{ID: domain.ID(id), BoardID: domain.ID(storedBoardID),
			Name: name, Color: color, RowVersion: version}
		if !label.Valid() {
			return nil, application.ErrUnavailable
		}
		result = append(result, label)
	}
	return result, rows.Err()
}
