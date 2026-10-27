package application

import (
	"context"
	"math"
	"strings"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) TraceListing(
	ctx context.Context,
	environmentID string,
	filter platformopsapi.TraceFilter,
) (platformopsapi.TracePage, error) {
	if service == nil || service.traces == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) || !validTraceFilter(filter) {
		return platformopsapi.TracePage{}, platformopsapi.ErrInvalid
	}
	result, err := service.traces.ListTraces(ctx, environmentID, filter)
	if err != nil {
		return platformopsapi.TracePage{}, mapRecoveryStoreError(err)
	}
	return result, nil
}

func validTraceFilter(filter platformopsapi.TraceFilter) bool {
	if filter.Page < 1 || filter.Limit < 1 || filter.Limit > 100 ||
		len(filter.ServiceName) > 100 || len(filter.Operation) > 200 ||
		len(filter.Search) > 200 || filter.MinimumDuration < 0 ||
		filter.MaximumDuration < 0 || (filter.MaximumDuration > 0 &&
		filter.MinimumDuration > filter.MaximumDuration) ||
		math.IsNaN(filter.MinimumDuration) || math.IsNaN(filter.MaximumDuration) ||
		math.IsInf(filter.MinimumDuration, 0) || math.IsInf(filter.MaximumDuration, 0) {
		return false
	}
	if filter.Level != "" && !validTraceLevel(filter.Level) {
		return false
	}
	if !validTraceStatus(filter.StatusCode) || !validTraceStatus(filter.MinStatus) ||
		!validTraceStatus(filter.MaxStatus) ||
		filter.MinStatus > 0 && filter.MaxStatus > 0 && filter.MinStatus > filter.MaxStatus {
		return false
	}
	if filter.StartedAfter != nil && filter.EndedBefore != nil &&
		(filter.EndedBefore.Before(*filter.StartedAfter) ||
			filter.EndedBefore.Sub(*filter.StartedAfter).Hours() > 24*31) {
		return false
	}
	if filter.SortBy != "started_at" && filter.SortBy != "duration_ms" &&
		filter.SortBy != "span_count" {
		return false
	}
	return strings.EqualFold(filter.SortDirection, "asc") ||
		strings.EqualFold(filter.SortDirection, "desc")
}

func validTraceStatus(value int) bool {
	return value == 0 || value >= 100 && value <= 599
}
