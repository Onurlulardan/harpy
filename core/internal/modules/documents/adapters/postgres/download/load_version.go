package download

import (
	"context"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) LoadVersion(ctx context.Context, itemID,
	versionID domain.ID,
) (downloadapp.Descriptor, error) {
	if !storage.usable(ctx) || !itemID.UUIDv7() || !versionID.UUIDv7() {
		return downloadapp.Descriptor{}, downloadapp.ErrInvalid
	}
	return scanDescriptor(storage.queryRow(ctx, getVersionSQL, itemID.String(),
		versionID.String()), storage.tenantID, storage.notFound)
}
