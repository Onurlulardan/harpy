package item

import (
	"context"

	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) List(ctx context.Context, folderID domain.ID,
	after itemapp.Cursor, limit uint16,
) (itemapp.Page, error) {
	if !storage.usable(ctx) || !folderID.UUIDv7() || limit == 0 || limit > 100 {
		return itemapp.Page{}, itemapp.ErrInvalid
	}
	var afterAt any
	var afterID any
	if after.ID.Valid() {
		afterAt, afterID = after.UpdatedAt, after.ID.String()
	}
	values, err := storage.query(ctx, listSQL, folderID.String(), afterAt,
		afterID, int(limit)+1)
	if err != nil {
		return itemapp.Page{}, itemapp.ErrUnavailable
	}
	defer values.Close()
	items := make([]itemapp.Snapshot, 0, limit)
	for values.Next() {
		var document []byte
		if err = values.Scan(&document); err != nil {
			return itemapp.Page{}, itemapp.ErrUnavailable
		}
		decoded, decodeErr := decodeItem(storage.tenantID, document)
		if decodeErr != nil {
			return itemapp.Page{}, decodeErr
		}
		items = append(items, decoded)
	}
	if values.Err() != nil {
		return itemapp.Page{}, itemapp.ErrUnavailable
	}
	page := itemapp.Page{Items: items}
	if len(items) > int(limit) {
		page.Items = items[:limit]
		last := page.Items[len(page.Items)-1]
		page.Next = itemapp.Cursor{UpdatedAt: last.UpdatedAt, ID: last.ID}
	}
	return page, nil
}
