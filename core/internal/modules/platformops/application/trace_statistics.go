package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) TraceStatistics(
	ctx context.Context,
	environmentID string,
) (platformopsapi.TraceStats, error) {
	if service == nil || service.traces == nil || ctx == nil || ctx.Err() != nil ||
		environmentID == "" {
		return platformopsapi.TraceStats{}, platformopsapi.ErrInvalid
	}
	result, err := service.traces.ReadTraceStats(ctx, environmentID)
	if err != nil {
		return platformopsapi.TraceStats{}, mapRecoveryStoreError(err)
	}
	return result, nil
}
