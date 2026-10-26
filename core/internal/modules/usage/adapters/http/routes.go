package usagehttp

import "net/http"

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	return []Route{
		{http.MethodGet, "/api/usage-analytics/overview", handler.overview(platformScope)},
		{http.MethodGet, "/api/usage-analytics/users/active", handler.activeUsers(platformScope)},
		{http.MethodGet, "/api/usage-analytics/users/top", handler.topUsers(platformScope)},
		{http.MethodGet, "/api/usage-analytics/users/{user_id}/summary", handler.userSummary(platformScope)},
		{http.MethodGet, "/api/usage-analytics/users/{user_id}/timeline", handler.userTimeline(platformScope)},
		{http.MethodGet, "/api/usage-analytics/features", handler.features(platformScope)},
		{http.MethodGet, "/api/usage-analytics/heatmap", handler.heatmap(platformScope)},
		{http.MethodGet, "/api/usage-analytics/trend", handler.trend(platformScope)},
		{http.MethodGet, "/api/usage-analytics/tenant/overview", handler.overview(tenantScope)},
		{http.MethodGet, "/api/usage-analytics/tenant/users/active", handler.activeUsers(tenantScope)},
		{http.MethodGet, "/api/usage-analytics/tenant/users/top", handler.topUsers(tenantScope)},
		{http.MethodGet, "/api/usage-analytics/tenant/users/{user_id}/summary", handler.userSummary(tenantScope)},
		{http.MethodGet, "/api/usage-analytics/tenant/users/{user_id}/timeline", handler.userTimeline(tenantScope)},
		{http.MethodGet, "/api/usage-analytics/tenant/features", handler.features(tenantScope)},
		{http.MethodGet, "/api/usage-analytics/tenant/heatmap", handler.heatmap(tenantScope)},
		{http.MethodGet, "/api/usage-analytics/tenant/trend", handler.trend(tenantScope)},
	}
}
