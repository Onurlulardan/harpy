package link

import (
	"context"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) LockActiveEpoch(
	ctx context.Context,
	itemID domain.ID,
) (uint64, error) {
	if !storage.usable(ctx) || !itemID.UUIDv7() {
		return 0, linkapp.ErrInvalid
	}
	var epoch int64
	if err := storage.queryRow(ctx, lockActiveEpochSQL, itemID.String()).Scan(
		&epoch,
	); err != nil {
		if storage.notFound(err) {
			return 0, linkapp.ErrNotFound
		}
		return 0, linkapp.ErrUnavailable
	}
	if epoch < 1 {
		return 0, linkapp.ErrUnavailable
	}
	return uint64(epoch), nil
}
