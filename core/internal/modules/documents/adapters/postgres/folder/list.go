package folder

import (
	"context"

	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) List(ctx context.Context, parent, after domain.ID,
	limit uint16,
) (folderapp.Page, error) {
	if !storage.usable(ctx) || parent.Valid() && !parent.UUIDv7() ||
		after.Valid() && !after.UUIDv7() || limit < 1 || limit > 200 {
		return folderapp.Page{}, folderapp.ErrInvalid
	}
	result, err := storage.query(ctx, listSQL, nullableID(parent), nullableID(after), limit+1)
	if err != nil {
		return folderapp.Page{}, folderapp.ErrUnavailable
	}
	defer result.Close()
	page := folderapp.Page{Folders: make([]domain.Folder, 0, limit)}
	for result.Next() {
		var document []byte
		if result.Scan(&document) != nil {
			return folderapp.Page{}, folderapp.ErrUnavailable
		}
		folder, decodeErr := decodeFolder(storage.tenantID, document)
		if decodeErr != nil {
			return folderapp.Page{}, decodeErr
		}
		if len(page.Folders) == int(limit) {
			page.NextID = page.Folders[len(page.Folders)-1].ID
			break
		}
		page.Folders = append(page.Folders, folder)
	}
	if result.Err() != nil {
		return folderapp.Page{}, folderapp.ErrUnavailable
	}
	return page, nil
}
