package download

import (
	"context"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) LoadCurrent(ctx context.Context,
	itemID domain.ID,
) (downloadapp.Descriptor, error) {
	if !storage.usable(ctx) || !itemID.UUIDv7() {
		return downloadapp.Descriptor{}, downloadapp.ErrInvalid
	}
	return scanDescriptor(storage.queryRow(ctx, getCurrentSQL, itemID.String()),
		storage.tenantID, storage.notFound)
}
