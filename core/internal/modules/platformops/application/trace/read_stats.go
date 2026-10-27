package trace

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *TraceOperations) ReadTraceStats(
	ctx context.Context, environmentID string,
) (platformopsapi.TraceStats, error) {
	return operations.backend.TraceStatistics(ctx, environmentID)
}
