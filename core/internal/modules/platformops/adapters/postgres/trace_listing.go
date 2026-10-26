package postgres

import (
	"context"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (store *Store) TraceListing(
	ctx context.Context,
	environmentID string,
	filter platformopsapi.TraceFilter,
) (platformopsapi.TracePage, error) {
	result := platformopsapi.TracePage{Traces: make([]platformopsapi.TraceSummary, 0),
		Page: filter.Page, Limit: filter.Limit}
	offset := (filter.Page - 1) * filter.Limit
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		rows, err := database.Query(ctx, listTracesSQL, environmentID,
			filter.ServiceName, filter.Operation, filter.Level, filter.StatusCode,
			filter.MinStatus, filter.MaxStatus, filter.MinimumDuration,
			filter.MaximumDuration, filter.StartedAfter, filter.EndedBefore,
			filter.Search, filter.SortBy, filter.SortDirection, filter.Limit, offset)
		if err != nil {
			return store.mapDatabaseError(err)
		}
		defer rows.Close()
		for rows.Next() {
			var item platformopsapi.TraceSummary
			var total int64
			if err = rows.Scan(&item.CorrelationID, &item.FirstSpan, &item.LastSpan,
				&item.TotalDurationMillis, &item.SpanCount, &item.Services,
				&item.RootOperation, &item.RootStatusCode, &item.HasErrors, &total); err != nil {
				return store.mapDatabaseError(err)
			}
			result.Total = total
			result.Traces = append(result.Traces, item)
		}
		return store.mapDatabaseError(rows.Err())
	})
	if err != nil {
		return platformopsapi.TracePage{}, fmt.Errorf("list platform traces: %w", err)
	}
	if result.Total > 0 {
		result.TotalPages = int((result.Total + int64(filter.Limit) - 1) / int64(filter.Limit))
	}
	return result, nil
}
