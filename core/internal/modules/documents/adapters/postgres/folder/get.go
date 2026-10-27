package folder

import (
	"context"

	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Get(ctx context.Context, id domain.ID) (domain.Folder, error) {
	if !storage.usable(ctx) || !id.UUIDv7() {
		return domain.Folder{}, folderapp.ErrInvalid
	}
	return scanFolder(storage.queryRow(ctx, getSQL, id.String()), storage.tenantID, storage.notFound)
}
