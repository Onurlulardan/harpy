package httpedge

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestattribution"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

const requestObservationTimeout = time.Second

func (handler *Handler) serveBuffered(
	w http.ResponseWriter,
	r *http.Request,
	correlationID requestmeta.CorrelationID,
) {
	startedAt := time.Now().UTC()
	var sourceIdentity [32]byte
	if _, err := io.ReadFull(rand.Reader, sourceIdentity[:]); err != nil {
		handler.report(fmt.Errorf("generate HTTP request observation identity: %w", err))
	}
	buffer := newResponseBuffer()
	if handler.invokeRoutes(buffer, r) {
		handler.report(fmt.Errorf("HTTP route panic recovered"))
		endedAt := time.Now().UTC()
		handler.writeFailure(w, r, correlationID, httpwire.FailureInternal)
		handler.observeRequest(r, correlationID, sourceIdentity,
			http.StatusInternalServerError, startedAt, endedAt)
		return
	}
	endedAt := time.Now().UTC()
	if err := buffer.flush(w); err != nil {
		handler.report(err)
	}
	handler.observeRequest(r, correlationID, sourceIdentity,
		buffer.statusCode(), startedAt, endedAt)
}

func (handler *Handler) observeRequest(
	r *http.Request,
	correlationID requestmeta.CorrelationID,
	sourceIdentity [32]byte,
	statusCode int,
	startedAt, endedAt time.Time,
) {
	pattern, found := requestmeta.RoutePatternFromContext(r.Context())
	if !found {
		return
	}
	attribution, attributionState := requestattribution.Snapshot(r.Context())
	observation := RequestObservation{CorrelationID: correlationID,
		SourceIdentity: sourceIdentity, Attribution: attribution,
		AttributionState: attributionState,
		Method:           r.Method, RoutePattern: pattern, StatusCode: statusCode,
		StartedAt: startedAt, EndedAt: endedAt}
	if !observation.valid() {
		handler.report(fmt.Errorf("HTTP request observation is invalid"))
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), requestObservationTimeout)
	defer cancel()
	if err := handler.observe(ctx, observation); err != nil {
		handler.report(fmt.Errorf("observe HTTP request: %w", err))
	}
}

func (handler *Handler) invokeRoutes(w http.ResponseWriter, r *http.Request) (panicked bool) {
	defer func() {
		panicked = recover() != nil
	}()
	handler.routes.ServeHTTP(w, r)
	return false
}
