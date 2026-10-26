package postgres

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func decodeUserSummary(document []byte, scope domain.Scope) (domain.UserSummary, error) {
	var wire summaryDocument
	if decodeDocument(document, &wire) != nil || len(wire.TopRoutes) > 10 ||
		len(wire.DailyBreakdown) > 32 || !validCount(wire.TotalRequests, wire.UniqueRoutes) ||
		!validTime(wire.LastActive) {
		return domain.UserSummary{}, domain.ErrUnavailable
	}
	if _, err := domain.NewTimelineQuery(scope, wire.SubjectID, 1, 1); err != nil {
		return domain.UserSummary{}, domain.ErrUnavailable
	}
	result := domain.UserSummary{SubjectID: wire.SubjectID, LastActive: wire.LastActive.UTC(),
		TotalRequests: wire.TotalRequests, UniqueRoutes: wire.UniqueRoutes,
		TopRoutes:      make([]domain.RouteUsage, 0, len(wire.TopRoutes)),
		DailyBreakdown: make([]domain.DailyActivity, 0, len(wire.DailyBreakdown))}
	for _, item := range wire.TopRoutes {
		if !domain.ValidRoutePattern(item.RoutePattern) || item.Method == "" ||
			!validCount(item.RequestCount) || !validMetric(item.AverageMillis) {
			return domain.UserSummary{}, domain.ErrUnavailable
		}
		result.TopRoutes = append(result.TopRoutes, domain.RouteUsage{
			RoutePattern: item.RoutePattern, Method: item.Method,
			RequestCount: item.RequestCount, AverageMillis: item.AverageMillis})
	}
	for _, item := range wire.DailyBreakdown {
		if _, err := time.Parse("2006-01-02", item.Date); err != nil || !validCount(item.RequestCount) {
			return domain.UserSummary{}, domain.ErrUnavailable
		}
		result.DailyBreakdown = append(result.DailyBreakdown,
			domain.DailyActivity{Date: item.Date, RequestCount: item.RequestCount})
	}
	return result, nil
}

func decodeTimeline(document []byte) (domain.TimelinePage, error) {
	var wire timelineDocument
	if decodeDocument(document, &wire) != nil || len(wire.Items) > 100 ||
		!validCount(wire.Total) || wire.Page < 1 || wire.Page > 1000 || wire.Limit < 1 ||
		wire.Limit > 100 || wire.TotalPages < 0 {
		return domain.TimelinePage{}, domain.ErrUnavailable
	}
	result := domain.TimelinePage{Items: make([]domain.TimelineEntry, 0, len(wire.Items)),
		Total: wire.Total, Page: wire.Page, Limit: wire.Limit, TotalPages: wire.TotalPages}
	for _, item := range wire.Items {
		if !validFactID(item.ID) || !domain.ValidRoutePattern(item.RoutePattern) ||
			item.Method == "" || item.StatusCode < 100 || item.StatusCode > 599 ||
			!validMetric(item.DurationMillis) || !validTime(item.OccurredAt) {
			return domain.TimelinePage{}, domain.ErrUnavailable
		}
		result.Items = append(result.Items, domain.TimelineEntry{ID: item.ID,
			RoutePattern: item.RoutePattern, Method: item.Method, StatusCode: item.StatusCode,
			DurationMillis: item.DurationMillis, OccurredAt: item.OccurredAt.UTC(),
			Level: timelineLevel(item.StatusCode)})
	}
	return result, nil
}
