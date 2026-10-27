package postgres

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	asyncobservability "github.com/Onurlulardan/harpy/core/internal/platform/async/observability"
	"github.com/jackc/pgx/v5"
)

type observabilityDatabase interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type observabilitySource struct {
	database observabilityDatabase
}

func (source *observabilitySource) Snapshot(
	ctx context.Context,
) (asyncobservability.Snapshot, error) {
	if ctx == nil || source == nil || source.database == nil {
		return asyncobservability.Snapshot{}, fmt.Errorf(
			"async observability source is unavailable",
		)
	}
	var outboxAge, outboxAlarm, jobLag, parkedJobs, parkedOutbox int64
	err := source.database.QueryRow(
		ctx, observeAsyncBacklogSQL, eventsv1.UndeliveredOutboxAlarmSeconds,
	).Scan(&outboxAge, &outboxAlarm, &jobLag, &parkedJobs, &parkedOutbox)
	if err != nil {
		return asyncobservability.Snapshot{}, fmt.Errorf("observe async backlog: %w", err)
	}
	snapshot, err := asyncobservability.NewSnapshot(
		outboxAge, outboxAlarm, jobLag, parkedJobs, parkedOutbox,
	)
	if err != nil {
		return asyncobservability.Snapshot{}, fmt.Errorf(
			"decode async backlog observation: %w", err,
		)
	}
	return snapshot, nil
}
