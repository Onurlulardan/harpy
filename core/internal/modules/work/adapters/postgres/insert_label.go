package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/insert_label.sql
var insertLabelSQL string

func (store *Store) InsertLabel(ctx context.Context, label domain.Label) error {
	if !label.Valid() {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, insertLabelSQL, label.ID, label.BoardID, label.Name, label.Color, label.RowVersion)
}
