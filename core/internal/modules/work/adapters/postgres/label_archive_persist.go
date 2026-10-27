package postgres

import (
	"context"
	_ "embed"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/label_archive_persist.sql
var persistLabelArchiveSQL string

func (store *Store) PersistLabelArchive(ctx context.Context, id domain.ID,
	expected int64, deletedAt *time.Time) error {
	if !id.Valid() || expected < 1 {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, persistLabelArchiveSQL, id, expected, deletedAt)
}
