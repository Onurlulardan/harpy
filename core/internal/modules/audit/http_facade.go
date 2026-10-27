package audit

import (
	"fmt"

	audithttp "github.com/Onurlulardan/harpy/core/internal/modules/audit/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/google/uuid"
)

type HTTPAccess = audithttp.Access
type HTTPAccessRequest = audithttp.AccessRequest
type HTTPAccessVerifier = audithttp.AccessVerifier
type HTTPApplication = audithttp.Application
type HTTPErrorSink = audithttp.ErrorSink
type HTTPCorrelationResolver = audithttp.CorrelationResolver

type HTTPHandlerDependencies struct {
	Access      HTTPAccessVerifier
	Application HTTPApplication
	Report      HTTPErrorSink
	Correlation HTTPCorrelationResolver
}

type HTTPHandler struct{ adapter *audithttp.Handler }

func NewHTTPAccess(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission Permission, operation string,
) (HTTPAccess, error) {
	return audithttp.NewAccess(principalID, tenantID, tokenVersion,
		authorizationVersion, authorizationClock, permission, operation)
}

func NewHTTPHandler(dependencies HTTPHandlerDependencies) (*HTTPHandler, error) {
	adapter, err := audithttp.NewHandler(audithttp.Dependencies{
		Access: dependencies.Access, Application: dependencies.Application,
		Report: dependencies.Report, Correlation: dependencies.Correlation})
	if err != nil {
		return nil, err
	}
	handler := &HTTPHandler{adapter: adapter}
	if !handler.Ready() {
		return nil, fmt.Errorf("audit HTTP handler is not ready")
	}
	return handler, nil
}

func (handler *HTTPHandler) Ready() bool {
	return handler != nil && handler.adapter != nil && handler.adapter.Ready() &&
		len(handler.adapter.Routes()) == 5
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
