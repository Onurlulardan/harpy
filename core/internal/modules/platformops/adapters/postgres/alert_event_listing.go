package postgres

import (
	"context"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (store *Store) AlertEventListing(
	ctx context.Context,
	environmentID string,
	filter platformopsapi.AlertEventFilter,
) (platformopsapi.AlertEventPage, error) {
	result := platformopsapi.AlertEventPage{Items: make([]platformopsapi.AlertEvent, 0),
		Page: filter.Page, Limit: filter.Limit}
	offset := (filter.Page - 1) * filter.Limit
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		rows, err := database.Query(ctx, listAlertEventsSQL, environmentID,
			filter.ServiceName, filter.Status, filter.Limit, offset)
		if err != nil {
			return store.mapDatabaseError(err)
		}
		defer rows.Close()
		for rows.Next() {
			item, scanErr := scanAlertEvent(rows, &result.Total)
			if scanErr != nil {
				return store.mapDatabaseError(scanErr)
			}
			result.Items = append(result.Items, item)
		}
		return store.mapDatabaseError(rows.Err())
	})
	if err != nil {
		return platformopsapi.AlertEventPage{}, fmt.Errorf("list alert events: %w", err)
	}
	if result.Total > 0 {
		result.TotalPages = int((result.Total + int64(filter.Limit) - 1) / int64(filter.Limit))
	}
	return result, nil
}
