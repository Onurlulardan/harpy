package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) TraceRecording(
	ctx context.Context,
	observation platformopsapi.TraceObservation,
) error {
	if service == nil || service.traces == nil || ctx == nil || ctx.Err() != nil ||
		!validTraceObservation(observation) {
		return platformopsapi.ErrInvalid
	}
	if err := service.traces.RecordTrace(ctx, observation); err != nil {
		return mapRecoveryStoreError(err)
	}
	return nil
}
