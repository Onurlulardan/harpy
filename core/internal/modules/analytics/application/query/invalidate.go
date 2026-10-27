package query

import (
	"context"

	"github.com/google/uuid"
)

func (application *Application) invalidate(ctx context.Context, queryID uuid.UUID) {
	if application != nil && !nilPort(application.cache) {
		_ = application.cache.Invalidate(ctx, queryID)
	}
}
