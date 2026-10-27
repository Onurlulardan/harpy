package adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

func (application *Application) BeginShadow(ctx context.Context,
	authority Authority,
) (Run, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || !validAuthority(authority) {
		return Run{}, ErrInvalid
	}
	now := application.clock.Now().UTC()
	if !foundation.ValidInstant(now) {
		return Run{}, ErrUnavailable
	}
	result, err := application.store.BeginShadow(ctx, authority, now)
	if err != nil {
		return Run{}, mapError(err)
	}
	if !validRun(result, authority) || result.State != "shadow" {
		return Run{}, ErrUnavailable
	}
	return result, nil
}
