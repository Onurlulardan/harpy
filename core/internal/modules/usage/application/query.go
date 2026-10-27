package application

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type TimeRangeRequest struct{ Start, End *time.Time }

func (service *Service) timeRange(request TimeRangeRequest) (domain.TimeRange, error) {
	if service == nil || service.clock == nil || (request.Start == nil) != (request.End == nil) {
		return domain.TimeRange{}, domain.ErrInvalid
	}
	now := service.clock.Now().UTC().Truncate(time.Microsecond)
	start, end := now.Add(-24*time.Hour), now
	if request.Start != nil {
		start, end = request.Start.UTC(), request.End.UTC()
	}
	if end.After(now.Add(time.Minute)) {
		return domain.TimeRange{}, domain.ErrInvalid
	}
	return domain.NewTimeRange(start, end)
}

func decorateOverview(value domain.Overview) domain.Overview {
	value.TopFeatures = decorateFeatures(value.TopFeatures)
	return value
}

func decorateFeatures(values []domain.FeatureUsage) []domain.FeatureUsage {
	if values == nil {
		return []domain.FeatureUsage{}
	}
	for index := range values {
		values[index].Feature = domain.FeatureForRoute(values[index].RoutePattern)
	}
	return values
}

func decorateRoutes(values []domain.RouteUsage) []domain.RouteUsage {
	if values == nil {
		return []domain.RouteUsage{}
	}
	for index := range values {
		values[index].Feature = domain.FeatureForRoute(values[index].RoutePattern)
	}
	return values
}

func decorateTimeline(value domain.TimelinePage) domain.TimelinePage {
	if value.Items == nil {
		value.Items = []domain.TimelineEntry{}
	}
	for index := range value.Items {
		value.Items[index].Feature = domain.FeatureForRoute(value.Items[index].RoutePattern)
	}
	return value
}
