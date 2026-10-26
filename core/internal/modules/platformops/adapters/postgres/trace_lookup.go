package postgres

import (
	"context"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) TraceLookup(
	ctx context.Context,
	environmentID, correlationID string,
) ([]platformopsapi.TraceSpan, error) {
	result := make([]platformopsapi.TraceSpan, 0)
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		rows, err := database.Query(ctx, getTraceSQL, environmentID, correlationID)
		if err != nil {
			return store.mapDatabaseError(err)
		}
		defer rows.Close()
		for rows.Next() {
			item, scanErr := scanTraceSpan(rows)
			if scanErr != nil {
				return store.mapDatabaseError(scanErr)
			}
			result = append(result, item)
		}
		return store.mapDatabaseError(rows.Err())
	})
	if err != nil {
		return nil, fmt.Errorf("get platform trace: %w", err)
	}
	if len(result) == 0 {
		return nil, ports.ErrNotFound
	}
	return result, nil
}
