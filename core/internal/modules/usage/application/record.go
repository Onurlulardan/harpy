package application

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) Record(ctx context.Context, request RecordRequest) (RecordResult, error) {
	if err := service.ready(ctx); err != nil {
		return RecordResult{}, domain.ErrUnavailable
	}
	id, err := service.ids.NewV7()
	if err != nil {
		return RecordResult{}, fmt.Errorf("generate usage fact identity: %w", domain.ErrUnavailable)
	}
	observation, err := domain.NormalizeObservation(domain.Observation{
		ID: domain.ID(id), EnvironmentID: domain.ID(request.EnvironmentID),
		TenantID: domain.ID(request.TenantID), ActorID: domain.ID(request.ActorID),
		SourceIdentity: request.SourceIdentity,
		Method:         request.Method, RoutePattern: request.RoutePattern,
		StatusCode: request.StatusCode, DurationMicros: request.DurationMicros,
		OccurredAt: request.OccurredAt,
	})
	if err != nil {
		return RecordResult{}, err
	}
	now := service.clock.Now().UTC()
	if now.IsZero() || observation.OccurredAt.After(now.Add(time.Minute)) ||
		observation.OccurredAt.Before(now.Add(-24*time.Hour)) {
		return RecordResult{}, domain.ErrInvalid
	}
	return service.store.Record(ctx, observation, observation.Digest())
}
