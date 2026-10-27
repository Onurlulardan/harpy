package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) TraceSpanLookup(
	ctx context.Context, environmentID, spanID string,
) (platformopsapi.TraceSpan, error) {
	if service == nil || service.traces == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) || !traceSpanIDPattern.MatchString(spanID) {
		return platformopsapi.TraceSpan{}, platformopsapi.ErrInvalid
	}
	result, err := service.traces.GetTraceSpan(ctx, environmentID, spanID)
	if err != nil {
		return platformopsapi.TraceSpan{}, mapRecoveryStoreError(err)
	}
	return result, nil
}
