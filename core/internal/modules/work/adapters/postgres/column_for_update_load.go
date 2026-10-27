package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_column_for_update.sql
var getColumnForUpdateSQL string

func (store *Store) LoadColumnForUpdate(ctx context.Context, id domain.ID) (domain.Column, error) {
	if !id.Valid() {
		return domain.Column{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.Column{}, err
	}
	var columnID, boardID uuid.UUID
	var name, color, status string
	var position, version int64
	err := store.database.QueryRow(ctx, getColumnForUpdateSQL, id).Scan(
		&columnID, &boardID, &name, &color, &status, &position, &version)
	if err != nil {
		return domain.Column{}, err
	}
	column := domain.Column{ID: domain.ID(columnID), BoardID: domain.ID(boardID),
		Name: name, Color: color, Status: domain.TaskStatus(status),
		Position: position, RowVersion: version}
	if !column.Valid() {
		return domain.Column{}, application.ErrUnavailable
	}
	return column, nil
}
