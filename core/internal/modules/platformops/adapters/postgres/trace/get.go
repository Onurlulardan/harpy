package trace

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *TraceOperations) GetTrace(
	ctx context.Context, environmentID, correlationID string,
) ([]platformopsapi.TraceSpan, error) {
	return operations.backend.TraceLookup(ctx, environmentID, correlationID)
}
