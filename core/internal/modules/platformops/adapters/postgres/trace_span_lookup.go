package postgres

import (
	"context"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (store *Store) TraceSpanLookup(
	ctx context.Context, environmentID, spanID string,
) (platformopsapi.TraceSpan, error) {
	var result platformopsapi.TraceSpan
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		value, err := scanTraceSpan(database.QueryRow(
			ctx, getTraceSpanSQL, environmentID, spanID))
		if err != nil {
			return store.mapDatabaseError(err)
		}
		result = value
		return nil
	})
	if err != nil {
		return platformopsapi.TraceSpan{}, fmt.Errorf("get platform trace span: %w", err)
	}
	return result, nil
}
