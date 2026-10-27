package usagehttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type routeScope struct {
	scope             domain.Scope
	overview, active  string
	top, summary      string
	timeline, feature string
	heatmap, trend    string
}

var tenantScope = routeScope{domain.TenantScope,
	application.OperationTenantOverview, application.OperationTenantActiveUsers,
	application.OperationTenantTopUsers, application.OperationTenantUserSummary,
	application.OperationTenantTimeline, application.OperationTenantFeatures,
	application.OperationTenantHeatmap, application.OperationTenantTrend}

var platformScope = routeScope{domain.PlatformScope,
	application.OperationPlatformOverview, application.OperationPlatformActiveUsers,
	application.OperationPlatformTopUsers, application.OperationPlatformUserSummary,
	application.OperationPlatformTimeline, application.OperationPlatformFeatures,
	application.OperationPlatformHeatmap, application.OperationPlatformTrend}

func (handler *Handler) overview(route routeScope) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		access, err := handler.accessFor(request, route.scope, route.overview)
		if err == nil {
			var window application.TimeRangeRequest
			window, err = timeRange(request)
			if err == nil {
				var result domain.Overview
				result, err = handler.dependencies.Application.Overview(request.Context(), access, window)
				if err == nil {
					handler.writeSuccess(response, request, overviewViewOf(result))
					return
				}
			}
		}
		handler.writeFailure(response, request, err)
	})
}

func (handler *Handler) activeUsers(route routeScope) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		access, err := handler.accessFor(request, route.scope, route.active)
		if err == nil && !onlyQueryKeys(request) {
			err = domain.ErrInvalid
		}
		if err == nil {
			var result domain.ActiveUsers
			result, err = handler.dependencies.Application.ActiveUsers(request.Context(), access)
			if err == nil {
				handler.writeSuccess(response, request, activeUsersViewOf(result))
				return
			}
		}
		handler.writeFailure(response, request, err)
	})
}

func (handler *Handler) topUsers(route routeScope) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		access, err := handler.accessFor(request, route.scope, route.top)
		window, limit := application.TimeRangeRequest{}, 0
		if err == nil {
			window, err = timeRange(request, "limit")
		}
		if err == nil {
			limit, err = positiveQuery(request, "limit", 10, 50)
		}
		if err == nil {
			var result []domain.TopUser
			result, err = handler.dependencies.Application.TopUsers(request.Context(), access, window, limit)
			if err == nil {
				handler.writeSuccess(response, request, topUsersViewOf(result))
				return
			}
		}
		handler.writeFailure(response, request, err)
	})
}

func (handler *Handler) userSummary(route routeScope) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		access, err := handler.accessFor(request, route.scope, route.summary)
		var window application.TimeRangeRequest
		if err == nil {
			window, err = timeRange(request)
		}
		if err == nil {
			var result domain.UserSummary
			result, err = handler.dependencies.Application.UserSummary(request.Context(), access,
				request.PathValue("user_id"), window)
			if err == nil {
				handler.writeSuccess(response, request, userSummaryViewOf(result))
				return
			}
		}
		handler.writeFailure(response, request, err)
	})
}

func (handler *Handler) userTimeline(route routeScope) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		access, err := handler.accessFor(request, route.scope, route.timeline)
		page, limit := 0, 0
		if err == nil && !onlyQueryKeys(request, "page", "limit") {
			err = domain.ErrInvalid
		}
		if err == nil {
			page, err = positiveQuery(request, "page", 1, 1000)
		}
		if err == nil {
			limit, err = positiveQuery(request, "limit", 20, 100)
		}
		if err == nil {
			var result domain.TimelinePage
			result, err = handler.dependencies.Application.UserTimeline(request.Context(), access,
				request.PathValue("user_id"), page, limit)
			if err == nil {
				handler.writeSuccess(response, request, timelinePageViewOf(result))
				return
			}
		}
		handler.writeFailure(response, request, err)
	})
}
