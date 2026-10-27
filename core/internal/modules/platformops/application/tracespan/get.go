package tracespan

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *TraceSpanOperations) GetTraceSpan(
	ctx context.Context, environmentID, spanID string,
) (platformopsapi.TraceSpan, error) {
	return operations.backend.TraceSpanLookup(ctx, environmentID, spanID)
}
