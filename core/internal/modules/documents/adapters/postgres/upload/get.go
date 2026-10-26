package upload

import (
	"context"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Get(ctx context.Context, id domain.ID) (uploadapp.Snapshot, error) {
	if !storage.usable(ctx) || !id.UUIDv7() {
		return uploadapp.Snapshot{}, uploadapp.ErrInvalid
	}
	return scanSnapshot(storage.queryRow(ctx, getSQL, id.String(), storage.actorID.String()),
		storage.tenantID, storage.notFound)
}
