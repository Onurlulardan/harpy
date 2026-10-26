package postgres

import (
	"context"
	"fmt"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/google/uuid"
)

func (store *Store) AlertTestStorage(
	ctx context.Context,
	environmentID, id, actorID, message string,
) (platformopsapi.AlertEvent, error) {
	var result platformopsapi.AlertEvent
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		row := database.QueryRow(ctx, testAlertRuleSQL, uuid.NewString(), id,
			environmentID, actorID, message, time.Now().UTC())
		var scanErr error
		result, scanErr = scanAlertEvent(row, nil)
		return store.mapDatabaseError(scanErr)
	})
	if err != nil {
		return platformopsapi.AlertEvent{}, fmt.Errorf("test alert rule: %w", err)
	}
	return result, nil
}
