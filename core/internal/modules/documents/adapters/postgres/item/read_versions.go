package item

import (
	"context"

	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) ReadVersions(ctx context.Context, itemID domain.ID,
	after uint64, limit uint16,
) (itemapp.VersionPage, error) {
	if !storage.usable(ctx) || !itemID.UUIDv7() || limit == 0 || limit > 100 {
		return itemapp.VersionPage{}, itemapp.ErrInvalid
	}
	result, err := storage.query(ctx, listVersionsSQL, itemID.String(), int64(after), limit)
	if err != nil {
		return itemapp.VersionPage{}, itemapp.ErrUnavailable
	}
	defer result.Close()
	page := itemapp.VersionPage{Items: make([]itemapp.Version, 0, limit)}
	for result.Next() {
		var version itemapp.Version
		var id string
		var number int64
		var digest []byte
		if err = result.Scan(&id, &number, &version.OriginalName, &version.MIME,
			&version.FileKind, &version.PlaintextBytes, &digest, &version.CreatedAt); err != nil {
			return itemapp.VersionPage{}, itemapp.ErrUnavailable
		}
		version.ID, err = domain.ParseID(id)
		if err != nil || number < 1 || len(digest) != len(version.PlaintextSHA256) {
			return itemapp.VersionPage{}, itemapp.ErrUnavailable
		}
		version.Number = uint64(number)
		copy(version.PlaintextSHA256[:], digest)
		page.Items = append(page.Items, version)
	}
	if err = result.Err(); err != nil {
		return itemapp.VersionPage{}, itemapp.ErrUnavailable
	}
	if len(page.Items) == int(limit) {
		page.NextAfterNumber = page.Items[len(page.Items)-1].Number
	}
	return page, nil
}
