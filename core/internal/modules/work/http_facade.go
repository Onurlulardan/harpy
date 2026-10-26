package work

import (
	"net/http"

	workhttp "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
	"github.com/google/uuid"
)

type HTTPAccess = workhttp.Access
type HTTPAccessRequest = workhttp.AccessRequest
type HTTPAccessVerifier = workhttp.AccessVerifier
type HTTPApplication = workhttp.Application
type HTTPListInput = workhttp.ListInput
type HTTPGetInput = workhttp.GetInput
type HTTPWriteInput = workhttp.WriteInput
type HTTPCompatibilityHit = workhttp.CompatibilityHit
type HTTPErrorSink = workhttp.ErrorSink

var (
	ErrHTTPAuthentication = workhttp.ErrAuthentication
	ErrHTTPForbidden      = workhttp.ErrForbidden
	ErrHTTPInvalid        = workhttp.ErrInvalid
	ErrHTTPNotFound       = workhttp.ErrNotFound
	ErrHTTPConflict       = workhttp.ErrConflict
	ErrHTTPUnavailable    = workhttp.ErrUnavailable
)

type HTTPHandlerDependencies struct {
	Access      HTTPAccessVerifier
	Application HTTPApplication
	Report      HTTPErrorSink
}

type HTTPHandler struct{ transport *workhttp.Handler }

func NewHTTPAccess(principalID, tenantID uuid.UUID, tokenVersion, authorizationVersion int64,
	authority, operation, consumerID, buildID string, compatibility bool,
) (HTTPAccess, error) {
	return workhttp.NewAccess(principalID, tenantID, tokenVersion, authorizationVersion,
		authority, operation, consumerID, buildID, compatibility)
}

func NewHTTPAccessWithEvidence(
	principalID, tenantID uuid.UUID, tokenVersion, authorizationVersion int64,
	authority, operation, consumerID, buildID string, compatibility, protectedAdmin bool,
) (HTTPAccess, error) {
	return workhttp.NewAccessWithEvidence(
		principalID, tenantID, tokenVersion, authorizationVersion,
		authority, operation, consumerID, buildID, compatibility, protectedAdmin,
	)
}

func NewHTTPHandler(dependencies HTTPHandlerDependencies) (*HTTPHandler, error) {
	transport, err := workhttp.NewHandler(workhttp.HandlerDependencies{
		Access:      dependencies.Access,
		Application: dependencies.Application, Report: dependencies.Report,
		Correlation: workCorrelation,
	})
	if err != nil {
		return nil, err
	}
	return &HTTPHandler{transport: transport}, nil
}

func (handler *HTTPHandler) Ready() bool {
	return handler != nil && handler.transport != nil && handler.transport.Ready()
}

func (handler *HTTPHandler) Routes() []httpedge.Route {
	if !handler.Ready() {
		return nil
	}
	routes := handler.transport.Routes()
	result := make([]httpedge.Route, len(routes))
	for index, route := range routes {
		result[index] = httpedge.Route{Method: route.Method,
			Pattern: route.Pattern, Handler: route.Handler}
	}
	return result
}

func workCorrelation(request *http.Request) string {
	if request != nil {
		if value, found := requestmeta.CorrelationIDFromContext(request.Context()); found {
			return value.String()
		}
	}
	return requestmeta.EntropyFailureCorrelationID.String()
}
