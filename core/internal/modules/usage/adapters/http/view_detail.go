package usagehttp

import "github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"

func userSummaryViewOf(value domain.UserSummary) userSummaryView {
	top := make([]pathUsageView, 0, len(value.TopRoutes))
	for _, route := range value.TopRoutes {
		top = append(top, pathUsageView{route.RoutePattern, route.Method, route.Feature,
			route.RequestCount, route.AverageMillis})
	}
	daily := make([]dailyActivityView, 0, len(value.DailyBreakdown))
	for _, item := range value.DailyBreakdown {
		daily = append(daily, dailyActivityView{item.Date, item.RequestCount})
	}
	return userSummaryView{value.SubjectID, value.TotalRequests, value.UniqueRoutes,
		value.LastActive, top, daily}
}

func timelinePageViewOf(value domain.TimelinePage) timelinePageView {
	items := make([]timelineEntryView, 0, len(value.Items))
	for _, item := range value.Items {
		items = append(items, timelineEntryView{item.ID, item.RoutePattern, item.Method,
			item.Feature, item.StatusCode, item.DurationMillis,
			item.OccurredAt, item.Level})
	}
	return timelinePageView{items, value.Total, value.Page, value.Limit, value.TotalPages}
}

func heatmapViewsOf(values []domain.HeatmapCell) []heatmapView {
	result := make([]heatmapView, 0, len(values))
	for _, value := range values {
		result = append(result, heatmapView{value.Day, value.Hour, value.Count})
	}
	return result
}
