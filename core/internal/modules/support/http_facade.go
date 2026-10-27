package support

import (
	"net/http"

	supporthttp "github.com/Onurlulardan/harpy/core/internal/modules/support/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
	"github.com/google/uuid"
)

type HTTPAccess = application.Access
type HTTPAccessScope = application.AccessScope
type HTTPAccessRequest = supporthttp.AccessRequest
type HTTPAccessVerifier = supporthttp.AccessVerifier
type HTTPRateKeyResolver = supporthttp.RateKeyResolver
type HTTPContactApplication = supporthttp.ContactApplication
type HTTPReporterBugApplication = supporthttp.ReporterBugApplication
type HTTPTenantBugApplication = supporthttp.TenantBugApplication
type HTTPPlatformBugApplication = supporthttp.PlatformBugApplication
type HTTPApplications = supporthttp.Applications
type HTTPBookingConfig = supporthttp.BookingConfig
type HTTPErrorSink = supporthttp.ErrorSink

const (
	HTTPReporterScope = application.ScopeReporter
	HTTPTenantScope   = application.ScopeTenant
	HTTPPlatformScope = application.ScopePlatform
)

var (
	ErrHTTPAuthentication = application.ErrAuthentication
	ErrHTTPForbidden      = application.ErrForbidden
	ErrHTTPUnavailable    = application.ErrUnavailable
)

type HTTPHandlerDependencies struct {
	Access       HTTPAccessVerifier
	Applications HTTPApplications
	RateKeys     HTTPRateKeyResolver
	Booking      HTTPBookingConfig
	Report       HTTPErrorSink
}

type HTTPHandler struct{ adapter *supporthttp.Handler }

func NewHTTPAccess(actorID, tenantID uuid.UUID, tokenVersion, authorizationVersion,
	authorityClock int64, scope HTTPAccessScope, permission, operation,
	reporterName, reporterEmail, tenantSlug, tenantName string,
) (HTTPAccess, error) {
	return application.NewAccess(actorID, tenantID, tokenVersion, authorizationVersion,
		authorityClock, scope, permission, operation, reporterName, reporterEmail,
		tenantSlug, tenantName)
}

func NewHTTPHandler(dependencies HTTPHandlerDependencies) (*HTTPHandler, error) {
	adapter, err := supporthttp.NewHandler(supporthttp.Dependencies{
		Access: dependencies.Access, Applications: dependencies.Applications,
		RateKeys: dependencies.RateKeys, Booking: dependencies.Booking,
		Report: dependencies.Report, Correlation: supportCorrelation,
	})
	if err != nil {
		return nil, err
	}
	return &HTTPHandler{adapter: adapter}, nil
}

func (handler *HTTPHandler) Ready() bool {
	return handler != nil && handler.adapter != nil && handler.adapter.Ready() &&
		len(handler.adapter.Routes()) == 17
}

func (handler *HTTPHandler) Routes() []httpedge.Route {
	if !handler.Ready() {
		return nil
	}
	routes := handler.adapter.Routes()
	result := make([]httpedge.Route, len(routes))
	for index, route := range routes {
		result[index] = httpedge.Route{Method: route.Method,
			Pattern: route.Pattern, Handler: route.Handler}
	}
	return result
}

func supportCorrelation(request *http.Request) string {
	if request != nil {
		if value, found := requestmeta.CorrelationIDFromContext(request.Context()); found {
			return value.String()
		}
	}
	return requestmeta.EntropyFailureCorrelationID.String()
}
