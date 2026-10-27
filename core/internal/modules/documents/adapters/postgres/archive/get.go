package archive

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Get(ctx context.Context,
	id domain.ID,
) (archiveapp.Snapshot, error) {
	if !storage.usable(ctx) || !id.UUIDv7() {
		return archiveapp.Snapshot{}, archiveapp.ErrInvalid
	}
	return storage.scanSnapshot(storage.queryRow(ctx, getSQL, id.String()))
}
