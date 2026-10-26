package item

import (
	"context"

	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Get(ctx context.Context, id domain.ID) (itemapp.Snapshot, error) {
	if !storage.usable(ctx) || !id.UUIDv7() {
		return itemapp.Snapshot{}, itemapp.ErrInvalid
	}
	return scanSnapshot(storage.queryRow(ctx, getSQL, id.String()),
		storage.tenantID, storage.notFound)
}
