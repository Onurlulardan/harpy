package publication

import (
	"context"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
)

func (storage *store) List(ctx context.Context, limit int, after *uuid.UUID) (publicationapp.Page, error) {
	if !storage.usable(ctx) {
		return publicationapp.Page{}, publicationapp.ErrUnavailable
	}
	if limit < 1 || limit > 100 || after != nil && !validUUIDv7(*after) {
		return publicationapp.Page{}, publicationapp.ErrInvalid
	}
	var cursor any
	if after != nil {
		cursor = *after
	}
	resultRows, err := storage.query(ctx, listSQL, limit, cursor)
	if err != nil {
		return publicationapp.Page{}, mapError(err, storage.notFound)
	}
	defer resultRows.Close()
	page := publicationapp.Page{Items: make([]publicationapp.Snapshot, 0, limit)}
	for resultRows.Next() {
		item, scanErr := scanSnapshot(resultRows, storage.notFound)
		if scanErr != nil {
			return publicationapp.Page{}, scanErr
		}
		page.Items = append(page.Items, item)
	}
	if err := resultRows.Err(); err != nil {
		return publicationapp.Page{}, mapError(err, storage.notFound)
	}
	if len(page.Items) == limit {
		cursor := page.Items[len(page.Items)-1].ReleaseID
		page.After = &cursor
	}
	return page, nil
}
