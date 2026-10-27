package httpedge

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestattribution"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type ErrorSink func(error)

type Handler struct {
	admission *Admission
	routes    http.Handler
	report    ErrorSink
	observe   RequestObserver
}

func NewHandler(
	admission *Admission,
	routes http.Handler,
	report ErrorSink,
	observe RequestObserver,
) (*Handler, error) {
	if admission == nil {
		return nil, fmt.Errorf("HTTP admission is required")
	}
	if routes == nil {
		return nil, fmt.Errorf("HTTP routes are required")
	}
	if report == nil {
		return nil, fmt.Errorf("HTTP error sink is required")
	}
	if observe == nil {
		return nil, fmt.Errorf("HTTP request observer is required")
	}
	return &Handler{admission: admission, routes: routes, report: report, observe: observe}, nil
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	correlationID, err := requestmeta.ResolveCorrelationID(r.Header)
	if err != nil {
		handler.report(err)
		correlationID = requestmeta.EntropyFailureCorrelationID
		handler.writeFailure(w, r, correlationID, httpwire.FailureInternal)
		return
	}
	r = r.WithContext(requestmeta.WithCorrelationID(r.Context(), correlationID))
	if _, admitted := handler.admission.Admit(); !admitted {
		handler.writeFailure(w, r, correlationID, httpwire.FailureServiceUnavailable)
		return
	}
	if !canonicalPath(r) {
		handler.writeFailure(w, r, correlationID, httpwire.FailureRouteNotFound)
		return
	}
	attributed, err := requestattribution.WithSlot(r.Context())
	if err != nil {
		handler.report(fmt.Errorf("prepare HTTP request attribution: %w", err))
	} else {
		r = r.WithContext(attributed)
	}
	handler.serveBuffered(w, r, correlationID)
}

func canonicalPath(r *http.Request) bool {
	value := r.URL.Path
	return value != "" && value[0] == '/' && r.URL.RawPath == "" &&
		!strings.Contains(value, "\\") && path.Clean(value) == value
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
