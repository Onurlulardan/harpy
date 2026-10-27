package usagehttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (handler *Handler) features(route routeScope) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		access, err := handler.accessFor(request, route.scope, route.feature)
		window, limit := application.TimeRangeRequest{}, 0
		if err == nil {
			window, err = timeRange(request, "limit")
		}
		if err == nil {
			limit, err = positiveQuery(request, "limit", 20, 100)
		}
		if err == nil {
			var result []domain.FeatureUsage
			result, err = handler.dependencies.Application.Features(request.Context(), access, window, limit)
			if err == nil {
				handler.writeSuccess(response, request, featureViewsOf(result))
				return
			}
		}
		handler.writeFailure(response, request, err)
	})
}

func (handler *Handler) heatmap(route routeScope) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		access, err := handler.accessFor(request, route.scope, route.heatmap)
		var window application.TimeRangeRequest
		if err == nil {
			window, err = timeRange(request)
		}
		if err == nil {
			var result []domain.HeatmapCell
			result, err = handler.dependencies.Application.Heatmap(request.Context(), access, window)
			if err == nil {
				handler.writeSuccess(response, request, heatmapViewsOf(result))
				return
			}
		}
		handler.writeFailure(response, request, err)
	})
}

func (handler *Handler) trend(route routeScope) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		access, err := handler.accessFor(request, route.scope, route.trend)
		var window application.TimeRangeRequest
		if err == nil {
			window, err = timeRange(request, "granularity")
		}
		granularity := domain.Granularity(request.URL.Query().Get("granularity"))
		if granularity == "" {
			granularity = domain.GranularityHour
		}
		if err == nil && !domain.ValidGranularity(granularity) {
			err = domain.ErrInvalid
		}
		if err == nil {
			var result []domain.TrendPoint
			result, err = handler.dependencies.Application.Trend(request.Context(), access,
				window, granularity)
			if err == nil {
				handler.writeSuccess(response, request, trendViewsOf(result))
				return
			}
		}
		handler.writeFailure(response, request, err)
	})
}
