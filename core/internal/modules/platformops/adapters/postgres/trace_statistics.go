package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (store *Store) TraceStatistics(
	ctx context.Context,
	environmentID string,
) (platformopsapi.TraceStats, error) {
	var result platformopsapi.TraceStats
	var services, levels, endpoints []byte
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		err := database.QueryRow(ctx, traceStatsSQL, environmentID).Scan(
			&result.TotalTraces, &result.TotalSpans, &result.ErrorRate,
			&result.AverageDurationMillis, &services, &levels,
			&result.TracesLast24Hours, &result.TracesLastHour, &endpoints)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return platformopsapi.TraceStats{}, fmt.Errorf("read platform trace stats: %w", err)
	}
	if err = json.Unmarshal(services, &result.SpansByService); err != nil {
		return platformopsapi.TraceStats{}, fmt.Errorf("decode trace service stats: %w", err)
	}
	if err = json.Unmarshal(levels, &result.SpansByLevel); err != nil {
		return platformopsapi.TraceStats{}, fmt.Errorf("decode trace level stats: %w", err)
	}
	if err = json.Unmarshal(endpoints, &result.SlowestEndpoints); err != nil {
		return platformopsapi.TraceStats{}, fmt.Errorf("decode endpoint stats: %w", err)
	}
	return result, nil
}
