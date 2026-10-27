package analytics

import (
	"context"
	"fmt"
	"reflect"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type HTTPGenerationResolver interface {
	ResolveTenantAnalyticsGenerations(context.Context, uuid.UUID) (int64, int64, error)
	ResolveTenantBillingGenerations(context.Context, uuid.UUID) (int64, int64, error)
}

type HTTPScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type HTTPRuntime struct {
	resolver     HTTPGenerationResolver
	coordinator  HTTPScopeCoordinator
	queries      *queryapp.Application
	dashboards   *dashboardapp.Application
	exports      *exportapp.Application
	publications *publicationapp.Application
}

func NewHTTPRuntime(
	resolver HTTPGenerationResolver, coordinator HTTPScopeCoordinator,
	queries *queryapp.Application, dashboards *dashboardapp.Application,
	publications *publicationapp.Application, exports *exportapp.Application,
) (*HTTPRuntime, error) {
	if nilHTTPPort(resolver) || nilHTTPPort(coordinator) || queries == nil ||
		dashboards == nil || publications == nil || exports == nil {
		return nil, fmt.Errorf("analytics HTTP runtime dependencies are invalid")
	}
	return &HTTPRuntime{resolver: resolver, coordinator: coordinator,
		queries: queries, dashboards: dashboards, publications: publications, exports: exports}, nil
}

func nilHTTPPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

var _ analyticshttp.Application = (*HTTPRuntime)(nil)
