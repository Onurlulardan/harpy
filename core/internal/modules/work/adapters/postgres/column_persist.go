package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/update_column.sql
var updateColumnSQL string

func (store *Store) PersistColumn(ctx context.Context, column domain.Column, expected int64) error {
	if !column.Valid() || expected < 1 || column.RowVersion != expected+1 {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, updateColumnSQL, column.ID, column.Name, column.Color,
		column.Status, column.Position, column.RowVersion, expected)
}
