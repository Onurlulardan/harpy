package adoption

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

func (application *Application) RecordCompatibilityHit(ctx context.Context,
	hit CompatibilityHit,
) (CompatibilityHit, error) {
	for _, value := range []string{hit.RouteContract, hit.ConsumerID, hit.BuildID} {
		if strings.TrimSpace(value) != value || value == "" || len(value) > 128 {
			return CompatibilityHit{}, ErrInvalid
		}
	}
	if application == nil || ctx == nil || ctx.Err() != nil || hit.HitCount != 0 ||
		!hit.FirstHitAt.IsZero() || !hit.LastHitAt.IsZero() {
		return CompatibilityHit{}, ErrInvalid
	}
	now := application.clock.Now().UTC()
	if !foundation.ValidInstant(now) {
		return CompatibilityHit{}, ErrUnavailable
	}
	hit.FirstHitAt, hit.LastHitAt = now, now
	result, err := application.store.RecordHit(ctx, hit)
	if err != nil {
		return CompatibilityHit{}, mapError(err)
	}
	if result.RouteContract != hit.RouteContract || result.ConsumerID != hit.ConsumerID ||
		result.BuildID != hit.BuildID || result.HitCount < 1 ||
		result.FirstHitAt.IsZero() || result.LastHitAt.Before(result.FirstHitAt) {
		return CompatibilityHit{}, ErrUnavailable
	}
	return result, nil
}
