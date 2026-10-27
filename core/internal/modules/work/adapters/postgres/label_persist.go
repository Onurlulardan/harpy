package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/update_label.sql
var updateLabelSQL string

func (store *Store) PersistLabel(ctx context.Context, label domain.Label, expected int64) error {
	if !label.Valid() || expected < 1 || label.RowVersion != expected+1 {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, updateLabelSQL, label.ID, label.BoardID, label.Name,
		label.Color, label.RowVersion, expected)
}
