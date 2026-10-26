package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/insert_column.sql
var insertColumnSQL string

func (store *Store) InsertColumn(ctx context.Context, column domain.Column) error {
	if !column.Valid() {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, insertColumnSQL, column.ID, column.BoardID, column.Name,
		column.Color, column.Status, column.Position, column.RowVersion)
}
