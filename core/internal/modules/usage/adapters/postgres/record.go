package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

func (store *Store) Record(ctx context.Context, observation domain.Observation,
	digest [32]byte,
) (application.RecordResult, error) {
	normalized, err := domain.NormalizeObservation(observation)
	if err != nil || digest == [32]byte{} || digest != normalized.Digest() {
		return application.RecordResult{}, domain.ErrInvalid
	}
	var result application.RecordResult
	err = store.within(ctx, CommandTransaction, func(database Database) error {
		return database.QueryRow(ctx, recordSQL, uuid.UUID(normalized.ID),
			uuid.UUID(normalized.EnvironmentID), uuid.UUID(normalized.TenantID),
			uuid.UUID(normalized.ActorID), normalized.SourceIdentity[:], digest[:],
			normalized.Method, normalized.RoutePattern, normalized.StatusCode,
			normalized.DurationMicros, normalized.OccurredAt).
			Scan(&result.FactID, &result.Replayed)
	})
	if err != nil {
		return application.RecordResult{}, store.normalize(err)
	}
	if result.FactID == [16]byte{} || result.FactID.Version() != 7 {
		return application.RecordResult{}, domain.ErrUnavailable
	}
	return result, nil
}

var _ application.Store = (*Store)(nil)
