package analytics

import (
	"fmt"
	"time"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	compatibilityhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/compatibility"
	dashboardhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/dashboard"
	dashboardexporthttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/dashboardexport"
	publicationhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/publication"
	queryhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/query"
	queryexecutionhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/queryexecution"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	"github.com/google/uuid"
)

type HTTPAccess = analyticshttp.Access
type HTTPAccessRequest = analyticshttp.AccessRequest
type HTTPAccessVerifier = analyticshttp.AccessVerifier
type HTTPLegacyApplication = analyticshttp.LegacyApplication
type HTTPPermission = analyticshttp.Permission
type HTTPRoute = analyticshttp.Route
type HTTPErrorSink = analyticshttp.ErrorSink
type HTTPCorrelationResolver = analyticshttp.CorrelationResolver
type HTTPTrustedClientIPResolver = analyticshttp.TrustedClientIPResolver

type HTTPHandler struct{ routes []HTTPRoute }

const analyticsHTTPRouteCount = 38

type HTTPHandlerDependencies struct {
	Access      HTTPAccessVerifier
	Application analyticshttp.Application
	Exports     dashboardexporthttp.Application
	Public      analyticshttp.PublicApplication
	Legacy      HTTPLegacyApplication
	Report      HTTPErrorSink
	Correlation HTTPCorrelationResolver
	ClientIP    HTTPTrustedClientIPResolver
}

func NewHTTPAccess(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time, permission HTTPPermission, operation string,
) (HTTPAccess, error) {
	return analyticshttp.NewAccess(principalID, tenantID, tokenVersion,
		authorizationVersion, authorizationClock, stepUpAuthenticatedAt, permission, operation)
}

func NewHTTPHandler(dependencies HTTPHandlerDependencies) (*HTTPHandler, error) {
	if httptransport.PortMissing(dependencies.Application) ||
		httptransport.PortMissing(dependencies.Public) || httptransport.PortMissing(dependencies.Legacy) {
		return nil, fmt.Errorf("analytics HTTP dependencies are required")
	}
	runtime, err := httptransport.NewRuntime(dependencies.Access, dependencies.Report,
		dependencies.Correlation, dependencies.ClientIP, analyticsHTTPFailure)
	if err != nil {
		return nil, err
	}
	queryHandler, err := queryhttp.NewHandler(runtime, dependencies.Application)
	if err != nil {
		return nil, err
	}
	queryExecutionHandler, err := queryexecutionhttp.NewHandler(runtime, dependencies.Application)
	if err != nil {
		return nil, err
	}
	dashboardHandler, err := dashboardhttp.NewHandler(runtime, dependencies.Application)
	if err != nil {
		return nil, err
	}
	exportHandler, err := dashboardexporthttp.NewHandler(runtime, dependencies.Exports)
	if err != nil {
		return nil, err
	}
	publicationHandler, err := publicationhttp.NewHandler(runtime, dependencies.Application, dependencies.Public)
	if err != nil {
		return nil, err
	}
	compatibilityHandler, err := compatibilityhttp.NewHandler(runtime, dependencies.Legacy, dependencies.Public)
	if err != nil {
		return nil, err
	}
	routes := append([]HTTPRoute{}, queryHandler.Routes()...)
	routes = append(routes, queryExecutionHandler.Routes()...)
	routes = append(routes, dashboardHandler.Routes()...)
	routes = append(routes, exportHandler.Routes()...)
	routes = append(routes, publicationHandler.Routes()...)
	routes = append(routes, compatibilityHandler.Routes()...)
	if err := validateAnalyticsRoutes(routes); err != nil {
		return nil, fmt.Errorf("analytics HTTP route manifest is invalid: %w", err)
	}
	return &HTTPHandler{routes: routes}, nil
}

func (handler *HTTPHandler) Routes() []HTTPRoute {
	if !handler.Ready() {
		return nil
	}
	return append([]HTTPRoute(nil), handler.routes...)
}

func (handler *HTTPHandler) Ready() bool {
	return handler != nil && validateAnalyticsRoutes(handler.routes) == nil
}

func validateAnalyticsRoutes(routes []HTTPRoute) error {
	if len(routes) != analyticsHTTPRouteCount {
		return fmt.Errorf("route count %d does not equal %d", len(routes), analyticsHTTPRouteCount)
	}
	seen := make(map[string]struct{}, len(routes))
	for index, route := range routes {
		if route.Method == "" || route.Pattern == "" || route.Handler == nil {
			return fmt.Errorf("route %d is incomplete", index)
		}
		key := route.Method + " " + route.Pattern
		if _, found := seen[key]; found {
			return fmt.Errorf("route %q is duplicated", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}
