package usage

import (
	"net/http"

	usagehttp "github.com/Onurlulardan/harpy/core/internal/modules/usage/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type HTTPAccessRequest = usagehttp.AccessRequest
type HTTPAccessVerifier = usagehttp.AccessVerifier
type HTTPApplication = usagehttp.Application
type HTTPErrorSink = usagehttp.ErrorSink

type HTTPHandlerDependencies struct {
	Access      HTTPAccessVerifier
	Application HTTPApplication
	Report      HTTPErrorSink
}

type HTTPHandler struct{ adapter *usagehttp.Handler }

func NewHTTPHandler(dependencies HTTPHandlerDependencies) (*HTTPHandler, error) {
	adapter, err := usagehttp.NewHandler(usagehttp.Dependencies{
		Access: dependencies.Access, Application: dependencies.Application,
		Report: dependencies.Report, Correlation: usageCorrelation})
	if err != nil {
		return nil, err
	}
	return &HTTPHandler{adapter: adapter}, nil
}

func (handler *HTTPHandler) Ready() bool {
	return handler != nil && handler.adapter != nil && handler.adapter.Ready() &&
		len(handler.adapter.Routes()) == 16
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

func usageCorrelation(request *http.Request) string {
	if request != nil {
		if value, found := requestmeta.CorrelationIDFromContext(request.Context()); found {
			return value.String()
		}
	}
	return requestmeta.EntropyFailureCorrelationID.String()
}
