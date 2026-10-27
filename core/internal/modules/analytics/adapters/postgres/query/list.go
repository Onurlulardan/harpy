package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func (storage *store) List(
	ctx context.Context, limit int, after *uuid.UUID,
) (queryapp.Page, error) {
	if !storage.usable(ctx) {
		return queryapp.Page{}, queryapp.ErrUnavailable
	}
	if limit < 1 || limit > 100 || after != nil && !validUUIDv7(*after) {
		return queryapp.Page{}, queryapp.ErrInvalid
	}
	var cursor any
	if after != nil {
		cursor = *after
	}
	resultRows, err := storage.query(ctx, listSQL, limit, cursor)
	if err != nil {
		return queryapp.Page{}, mapError(err, storage.notFound)
	}
	defer resultRows.Close()
	page := queryapp.Page{Items: make([]queryapp.Snapshot, 0, limit)}
	for resultRows.Next() {
		item, scanErr := scanSnapshot(resultRows, storage.notFound)
		if scanErr != nil {
			return queryapp.Page{}, scanErr
		}
		page.Items = append(page.Items, item)
	}
	if err := resultRows.Err(); err != nil {
		return queryapp.Page{}, mapError(err, storage.notFound)
	}
	if len(page.Items) == limit {
		cursor := page.Items[len(page.Items)-1].QueryID
		page.After = &cursor
	}
	return page, nil
}
