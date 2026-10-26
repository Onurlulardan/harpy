package archive

import (
	"context"
	"time"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Expire(ctx context.Context, id, deletionID domain.ID,
	now time.Time,
) (archiveapp.Snapshot, bool, error) {
	if !storage.usable(ctx) || !storage.job || !storage.command || !id.UUIDv7() ||
		!deletionID.UUIDv7() || now.IsZero() {
		return archiveapp.Snapshot{}, false, archiveapp.ErrInvalid
	}
	var changedID string
	err := storage.queryRow(ctx, expireSQL, id.String(), deletionID.String(),
		now.UTC()).Scan(&changedID)
	if err == nil {
		result, getErr := storage.Get(ctx, id)
		return result, true, getErr
	}
	if !storage.notFound(err) {
		return archiveapp.Snapshot{}, false, archiveapp.ErrUnavailable
	}
	result, getErr := storage.Get(ctx, id)
	if getErr == nil && result.State == domain.ArchiveExpired {
		return result, false, nil
	}
	return archiveapp.Snapshot{}, false, archiveapp.ErrConflict
}
