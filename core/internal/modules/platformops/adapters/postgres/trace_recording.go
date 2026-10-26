package postgres

import (
	"context"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/google/uuid"
)

func (store *Store) TraceRecording(
	ctx context.Context,
	observation platformopsapi.TraceObservation,
) error {
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		_, err := database.Execute(ctx, recordTraceSQL, uuid.NewString(),
			observation.EnvironmentID, observation.CorrelationID, observation.SpanID,
			nullableText(observation.ParentSpanID), observation.ServiceName,
			observation.Operation, observation.SpanKind, observation.Method,
			observation.Path, observation.StatusCode, observation.StartedAt,
			observation.EndedAt, observation.DurationMillis, observation.Level,
			nullableText(observation.Message), observation.EndedAt)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("record platform operation trace: %w", err)
	}
	return nil
}
