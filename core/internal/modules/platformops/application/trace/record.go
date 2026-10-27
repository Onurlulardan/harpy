package trace

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *TraceOperations) RecordTrace(
	ctx context.Context, observation platformopsapi.TraceObservation,
) error {
	return operations.backend.TraceRecording(ctx, observation)
}
