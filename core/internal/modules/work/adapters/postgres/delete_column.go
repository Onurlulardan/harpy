package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/delete_empty_column.sql
var deleteEmptyColumnSQL string

func (store *Store) DeleteColumn(ctx context.Context, id domain.ID, expected int64) error {
	if !id.Valid() || expected < 1 {
		return application.ErrInvalid
	}
	return executeOne(ctx, store, deleteEmptyColumnSQL, id, expected)
}
