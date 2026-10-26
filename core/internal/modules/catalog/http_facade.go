package catalog

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	dependencyhttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http/dependency"
	lifecyclehttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http/lifecycle"
	menuhttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http/menu"
	registryhttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http/registry"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
	"github.com/google/uuid"
)

var (
	ErrHTTPAuthentication = cataloghttp.ErrAuthentication
	ErrHTTPForbidden      = cataloghttp.ErrForbidden
	ErrHTTPInvalid        = cataloghttp.ErrInvalid
	ErrHTTPNotFound       = cataloghttp.ErrNotFound
	ErrHTTPConflict       = cataloghttp.ErrConflict
	ErrHTTPUnavailable    = cataloghttp.ErrUnavailable
)

type HTTPAccess = cataloghttp.Access
type HTTPAccessRequest = cataloghttp.AccessRequest
type HTTPAccessVerifier = cataloghttp.AccessVerifier
type HTTPPermission = cataloghttp.Permission
type HTTPErrorSink = cataloghttp.ErrorSink

type HTTPHandler struct {
	runtime *cataloghttp.Runtime
	routes  []cataloghttp.Route
}

func NewHTTPAccessWithAuthorizationEvidence(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time,
	authorizationPermissionID uuid.UUID,
	authorizationEvidenceSHA256, stepUpPurposeSHA256 [sha256.Size]byte,
	permission HTTPPermission,
	operation string,
) (HTTPAccess, error) {
	return cataloghttp.NewAccessWithAuthorizationEvidence(
		principalID, tenantID, tokenVersion, authorizationVersion,
		authorizationClock, stepUpAuthenticatedAt, authorizationPermissionID,
		authorizationEvidenceSHA256, stepUpPurposeSHA256, permission, operation,
	)
}

type HTTPHandlerDependencies struct {
	Access      HTTPAccessVerifier
	Application cataloghttp.Application
	Report      HTTPErrorSink
}

func NewHTTPAccess(principalID, tenantID uuid.UUID, tokenVersion, authorizationVersion,
	authorizationClock int64, stepUpAuthenticatedAt time.Time,
	permission HTTPPermission, operation string,
) (HTTPAccess, error) {
	return cataloghttp.NewAccess(principalID, tenantID, tokenVersion, authorizationVersion,
		authorizationClock, stepUpAuthenticatedAt, permission, operation)
}

func NewHTTPHandler(dependencies HTTPHandlerDependencies) (*HTTPHandler, error) {
	if !cataloghttp.PortReady(dependencies.Application) {
		return nil, fmt.Errorf("catalog HTTP application is required")
	}
	runtime, err := cataloghttp.NewRuntime(dependencies.Access, dependencies.Report,
		catalogCorrelation)
	if err != nil {
		return nil, err
	}
	groups := [][]cataloghttp.Route{
		registryhttp.New(runtime, dependencies.Application).Routes(),
		dependencyhttp.New(runtime, dependencies.Application).Routes(),
		lifecyclehttp.New(runtime, dependencies.Application).Routes(),
		menuhttp.New(runtime, dependencies.Application).Routes(),
		cataloghttp.ModuleMigrationRetirementRoutes(runtime),
	}
	routes := make([]cataloghttp.Route, 0, 32)
	for _, group := range groups {
		routes = append(routes, group...)
	}
	if err := validateCatalogRoutes(routes); err != nil {
		return nil, fmt.Errorf("catalog HTTP route manifest is invalid: %w", err)
	}
	handler := &HTTPHandler{runtime: runtime, routes: routes}
	if !handler.Ready() {
		return nil, fmt.Errorf("catalog HTTP handler is not ready")
	}
	return handler, nil
}

func (handler *HTTPHandler) Ready() bool {
	return handler != nil && handler.runtime != nil && handler.runtime.Ready() &&
		validateCatalogRoutes(handler.routes) == nil
}

func validateCatalogRoutes(routes []cataloghttp.Route) error {
	if len(routes) != 32 {
		return fmt.Errorf("route count %d does not equal 32", len(routes))
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

func (handler *HTTPHandler) Routes() []httpedge.Route {
	if !handler.Ready() {
		return nil
	}
	result := make([]httpedge.Route, len(handler.routes))
	for index, route := range handler.routes {
		result[index] = httpedge.Route{Method: route.Method, Pattern: route.Pattern,
			Handler: route.Handler}
	}
	return result
}

func catalogCorrelation(request *http.Request) string {
	if request != nil {
		if value, found := requestmeta.CorrelationIDFromContext(request.Context()); found {
			return value.String()
		}
	}
	return requestmeta.EntropyFailureCorrelationID.String()
}
