package link

import (
	"context"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) ListBlockers(ctx context.Context,
	query linkapp.BlockerQuery,
) ([]domain.Link, error) {
	if !storage.usable(ctx) || !query.ItemID.UUIDv7() || query.Limit < 1 ||
		query.Limit > linkapp.MaximumBlockerPageSize+1 ||
		(query.AfterLinkID.Valid() && !query.AfterLinkID.UUIDv7()) {
		return nil, linkapp.ErrInvalid
	}
	var after any
	if query.AfterLinkID.Valid() {
		after = query.AfterLinkID.String()
	}
	result, err := storage.query(ctx, listSQL, query.ItemID.String(), after, query.Limit)
	if err != nil {
		return nil, linkapp.ErrUnavailable
	}
	defer result.Close()
	items := make([]domain.Link, 0, query.Limit)
	for result.Next() {
		item, _, scanErr := storage.scan(result)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := result.Err(); err != nil {
		return nil, linkapp.ErrUnavailable
	}
	return items, nil
}
