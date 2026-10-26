package documents

import (
	"fmt"
	"net/http"
	"time"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
	"github.com/google/uuid"
)

var (
	ErrHTTPAuthentication        = documentshttp.ErrAuthentication
	ErrHTTPForbidden             = documentshttp.ErrForbidden
	ErrHTTPInvalid               = documentshttp.ErrInvalid
	ErrHTTPNotFound              = documentshttp.ErrNotFound
	ErrHTTPConflict              = documentshttp.ErrConflict
	ErrHTTPContentRejected       = documentshttp.ErrContentRejected
	ErrHTTPScanUnavailable       = documentshttp.ErrScanUnavailable
	ErrHTTPClientUpgradeRequired = documentshttp.ErrClientUpgradeRequired
	ErrHTTPUnavailable           = documentshttp.ErrUnavailable
)

type HTTPAccess = documentshttp.Access
type HTTPAccessRequest = documentshttp.AccessRequest
type HTTPAccessVerifier = documentshttp.AccessVerifier
type HTTPErrorSink = documentshttp.ErrorSink

type HTTPHandler struct{ adapter *documentshttp.Handler }

type HTTPHandlerDependencies struct {
	Access      HTTPAccessVerifier
	Application documentshttp.Application
	Report      HTTPErrorSink
}

func NewHTTPAccess(principalID, tenantID uuid.UUID, tokenVersion,
	authorizationVersion, authorizationClock int64, stepUpAuthenticatedAt time.Time,
	permission, operation string,
) (HTTPAccess, error) {
	return documentshttp.NewAccessWithStepUp(principalID, tenantID, tokenVersion,
		authorizationVersion, authorizationClock, stepUpAuthenticatedAt, permission, operation)
}

func NewHTTPHandler(dependencies HTTPHandlerDependencies) (*HTTPHandler, error) {
	adapter, err := documentshttp.NewHandler(documentshttp.Dependencies{
		Access: dependencies.Access, Application: dependencies.Application,
		Report: dependencies.Report, Correlation: documentsCorrelation,
	})
	if err != nil {
		return nil, err
	}
	handler := &HTTPHandler{adapter: adapter}
	if !handler.Ready() {
		return nil, fmt.Errorf("Documents HTTP handler is not ready")
	}
	return handler, nil
}

func (handler *HTTPHandler) Ready() bool {
	return handler != nil && handler.adapter != nil && len(handler.adapter.Routes()) == 32
}

func (handler *HTTPHandler) Routes() []httpedge.Route {
	if !handler.Ready() {
		return nil
	}
	routes := handler.adapter.Routes()
	result := make([]httpedge.Route, len(routes))
	for index, route := range routes {
		result[index] = httpedge.Route{Method: route.Method, Pattern: route.Pattern,
			Handler: route.Handler}
	}
	return result
}

func documentsCorrelation(request *http.Request) string {
	if request != nil {
		if value, found := requestmeta.CorrelationIDFromContext(request.Context()); found {
			return value.String()
		}
	}
	return requestmeta.EntropyFailureCorrelationID.String()
}
