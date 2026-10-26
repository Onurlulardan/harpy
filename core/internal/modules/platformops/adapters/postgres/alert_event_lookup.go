package postgres

import (
	"context"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (store *Store) AlertEventLookup(
	ctx context.Context, environmentID, eventID string,
) (platformopsapi.AlertEvent, error) {
	var result platformopsapi.AlertEvent
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		value, err := scanAlertEvent(database.QueryRow(
			ctx, getAlertEventSQL, environmentID, eventID), nil)
		if err != nil {
			return store.mapDatabaseError(err)
		}
		result = value
		return nil
	})
	if err != nil {
		return platformopsapi.AlertEvent{}, fmt.Errorf("get alert event: %w", err)
	}
	return result, nil
}
