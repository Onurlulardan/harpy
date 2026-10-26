package analytics

import (
	"context"
	"encoding/json"
	"time"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type publicDashboardRoot interface {
	DashboardRoot(context.Context, platformtx.TenantScope,
		uuid.UUID, uuid.UUID) (uuid.UUID, error)
}

type publicDashboardSource interface {
	LoadPublishedVersion(context.Context, platformtx.TenantScope,
		uuid.UUID, uuid.UUID) (dashboardapp.Snapshot, error)
}

type publicExecutor struct {
	queries    *queryapp.PublicApplication
	roots      publicDashboardRoot
	dashboards publicDashboardSource
}

func newPublicExecutor(
	queries *queryapp.PublicApplication, roots publicDashboardRoot,
	dashboards publicDashboardSource,
) (*publicExecutor, error) {
	if queries == nil || nilHTTPPort(roots) || nilHTTPPort(dashboards) {
		return nil, publicationapp.ErrInvalid
	}
	return &publicExecutor{queries: queries, roots: roots, dashboards: dashboards}, nil
}

func (executor *publicExecutor) Execute(
	ctx context.Context, scope platformtx.TenantScope,
	route publicationapp.PublicRoute, parameters []byte,
) (publicationapp.PublicResult, error) {
	if executor == nil || executor.queries == nil || ctx == nil || ctx.Err() != nil ||
		route.ReleaseID == uuid.Nil || route.ResourceVersionID == uuid.Nil {
		return publicationapp.PublicResult{}, publicationapp.ErrUnavailable
	}
	timed, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	switch route.ResourceKind {
	case publicationapp.ResourceQuery:
		result, err := executor.queries.Run(
			timed, scope, route.ReleaseID, route.ResourceVersionID, parameters)
		if err != nil || len(result.Body) == 0 || len(result.Body) > 1<<20 {
			return publicationapp.PublicResult{}, publicationapp.ErrUnavailable
		}
		return publicationapp.PublicResult{Body: result.Body}, nil
	case publicationapp.ResourceDashboard:
		return executor.executeDashboard(timed, scope, route, parameters)
	default:
		return publicationapp.PublicResult{}, publicationapp.ErrUnavailable
	}
}

type publicWidgetResult struct {
	WidgetID string          `json:"widget_id"`
	Body     json.RawMessage `json:"body"`
}

type publicDashboardResult struct {
	Widgets []publicWidgetResult `json:"widgets"`
}

var _ publicationapp.PublicExecutor = (*publicExecutor)(nil)
