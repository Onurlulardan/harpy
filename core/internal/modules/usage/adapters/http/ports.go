package usagehttp

import (
	"context"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type Application interface {
	Overview(context.Context, application.Access, application.TimeRangeRequest) (domain.Overview, error)
	ActiveUsers(context.Context, application.Access) (domain.ActiveUsers, error)
	TopUsers(context.Context, application.Access, application.TimeRangeRequest, int) ([]domain.TopUser, error)
	UserSummary(context.Context, application.Access, string, application.TimeRangeRequest) (domain.UserSummary, error)
	UserTimeline(context.Context, application.Access, string, int, int) (domain.TimelinePage, error)
	Features(context.Context, application.Access, application.TimeRangeRequest, int) ([]domain.FeatureUsage, error)
	Heatmap(context.Context, application.Access, application.TimeRangeRequest) ([]domain.HeatmapCell, error)
	Trend(context.Context, application.Access, application.TimeRangeRequest, domain.Granularity) ([]domain.TrendPoint, error)
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (application.Access, error)
}

type ErrorSink func(error)
type CorrelationResolver func(*http.Request) string

type Route struct {
	Method, Pattern string
	Handler         http.Handler
}
