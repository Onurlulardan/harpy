package trace

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *TraceOperations) ListTraces(
	ctx context.Context, environmentID string, filter platformopsapi.TraceFilter,
) (platformopsapi.TracePage, error) {
	return operations.backend.TraceListing(ctx, environmentID, filter)
}
