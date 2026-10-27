package management

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type ErrorSink func(error)

type Handler struct {
	monitor *health.Monitor
	info    buildinfo.Info
	report  ErrorSink
}

func NewHandler(monitor *health.Monitor, info buildinfo.Info, report ErrorSink) (*Handler, error) {
	if monitor == nil {
		return nil, fmt.Errorf("management health monitor is required")
	}
	if report == nil {
		return nil, fmt.Errorf("management error sink is required")
	}
	if monitor.Service() != info.Service() {
		return nil, fmt.Errorf("management service identities do not match")
	}
	return &Handler{monitor: monitor, info: info, report: report}, nil
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	correlationID, err := requestmeta.ResolveCorrelationID(r.Header)
	if err != nil {
		handler.report(err)
		correlationID = requestmeta.EntropyFailureCorrelationID
		handler.writeFailure(w, r, correlationID, httpwire.FailureInternal)
		return
	}
	switch r.URL.Path {
	case "/livez", "/readyz", "/version":
	default:
		handler.writeFailure(w, r, correlationID, httpwire.FailureRouteNotFound)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		handler.writeFailure(w, r, correlationID, httpwire.FailureMethodNotAllowed)
		return
	}
	switch r.URL.Path {
	case "/livez":
		handler.writeLivez(w, r, correlationID)
	case "/readyz":
		handler.writeReadyz(w, r, correlationID)
	case "/version":
		handler.writeVersion(w, r, correlationID)
	}
}

func (handler *Handler) writeFailure(
	w http.ResponseWriter,
	r *http.Request,
	correlationID requestmeta.CorrelationID,
	failure httpwire.Failure,
) {
	if err := httpwire.WriteFailure(w, correlationID, failure, r.Method == http.MethodHead); err != nil {
		handler.report(err)
	}
}
