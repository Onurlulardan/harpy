package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) TraceLookup(
	ctx context.Context,
	environmentID, correlationID string,
) ([]platformopsapi.TraceSpan, error) {
	if service == nil || service.traces == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) || !validCorrelationID(correlationID) {
		return nil, platformopsapi.ErrInvalid
	}
	result, err := service.traces.GetTrace(ctx, environmentID, correlationID)
	if err != nil {
		return nil, mapRecoveryStoreError(err)
	}
	return result, nil
}
