package adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

func (application *Application) CreateRun(ctx context.Context,
	authority Authority,
) (Run, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || !validAuthority(authority) {
		return Run{}, ErrInvalid
	}
	now := application.clock.Now().UTC()
	if !foundation.ValidInstant(now) {
		return Run{}, ErrUnavailable
	}
	result, _, err := application.store.CreateRun(ctx, authority, TransformVersion, now)
	if err != nil {
		return Run{}, mapError(err)
	}
	if !validRun(result, authority) || result.State != "inventory" {
		return Run{}, ErrUnavailable
	}
	return result, nil
}
