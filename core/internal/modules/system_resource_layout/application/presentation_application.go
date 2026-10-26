package application

import (
	"context"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

type PresentationApplication struct {
	store Store
	clock Clock
}

func NewPresentationApplication(
	store Store,
	clock Clock,
) (*PresentationApplication, error) {
	if store == nil || clock == nil {
		return nil, layoutapi.ErrInvalid
	}
	return &PresentationApplication{store: store, clock: clock}, nil
}

func (application *PresentationApplication) usable(ctx context.Context) bool {
	return application != nil && application.store != nil && application.clock != nil &&
		ctx != nil && ctx.Err() == nil
}

func (application *PresentationApplication) currentTime() (time.Time, error) {
	value := application.clock.Now().UTC().Truncate(time.Microsecond)
	if value.IsZero() {
		return time.Time{}, layoutapi.ErrUnavailable
	}
	return value, nil
}
