package application

import (
	"context"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func (application *LayoutApplication) usable(ctx context.Context) bool {
	return application != nil && application.store != nil &&
		application.ids != nil && application.clock != nil &&
		ctx != nil && ctx.Err() == nil
}

func (application *LayoutApplication) currentTime() (time.Time, error) {
	value := application.clock.Now().UTC()
	if value.IsZero() {
		return time.Time{}, layoutapi.ErrUnavailable
	}
	return value, nil
}

func cloneRevision(input layoutapi.Revision) layoutapi.Revision {
	input.Definition = normalizeDefinition(input.Definition)
	input.PublishedAt = cloneTime(input.PublishedAt)
	return input
}
