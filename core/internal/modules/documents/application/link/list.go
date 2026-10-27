package link

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const MaximumBlockerPageSize = 100

func (application *Application) ListReferenceBlockers(ctx context.Context,
	scope platformtx.TenantScope, query BlockerQuery,
) (BlockerPage, error) {
	if application == nil || scope.OperationID() != "documents.link.list" ||
		!query.ItemID.UUIDv7() || query.Limit < 1 || query.Limit > MaximumBlockerPageSize ||
		(query.AfterLinkID.Valid() && !query.AfterLinkID.UUIDv7()) {
		return BlockerPage{}, ErrInvalid
	}
	store, identity, err := application.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return BlockerPage{}, ErrUnavailable
	}
	query.Limit++
	items, err := store.ListBlockers(ctx, query)
	if err != nil {
		return BlockerPage{}, mapError(err)
	}
	for _, item := range items {
		if item.TenantID != identity.TenantID || item.ItemID != query.ItemID ||
			item.State != domain.LinkActive || item.Usage != "LIFECYCLE" {
			return BlockerPage{}, ErrUnavailable
		}
	}
	page := BlockerPage{Items: items}
	if len(items) == int(query.Limit) {
		page.Items = items[:len(items)-1]
		page.NextLinkID = page.Items[len(page.Items)-1].ID
	}
	return page, nil
}
