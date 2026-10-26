package dashboard

import (
	"context"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"github.com/google/uuid"
)

func (storage *store) List(ctx context.Context, limit int, after *uuid.UUID) (dashboardapp.Page, error) {
	if !storage.usable(ctx) {
		return dashboardapp.Page{}, dashboardapp.ErrUnavailable
	}
	if limit < 1 || limit > 100 || after != nil && !validUUIDv7(*after) {
		return dashboardapp.Page{}, dashboardapp.ErrInvalid
	}
	var cursor any
	if after != nil {
		cursor = *after
	}
	resultRows, err := storage.query(ctx, listSQL, limit, cursor)
	if err != nil {
		return dashboardapp.Page{}, mapError(err, storage.notFound)
	}
	defer resultRows.Close()
	page := dashboardapp.Page{Items: make([]dashboardapp.Snapshot, 0, limit)}
	for resultRows.Next() {
		item, scanErr := scanSnapshot(resultRows, storage.notFound)
		if scanErr != nil {
			return dashboardapp.Page{}, scanErr
		}
		page.Items = append(page.Items, item)
	}
	if err := resultRows.Err(); err != nil {
		return dashboardapp.Page{}, mapError(err, storage.notFound)
	}
	if len(page.Items) == limit {
		cursor := page.Items[len(page.Items)-1].DashboardID
		page.After = &cursor
	}
	return page, nil
}
