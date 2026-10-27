package httpedge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func TestClosedEdgeRejectsBeforeRoutes(t *testing.T) {
	called := false
	handler := newTestHandler(t, &Admission{}, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}), nil)
	recorder := requestEdge(handler, http.MethodGet, "/api/resource")
	assertEdgeFailure(t, recorder, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE")
	if called {
		t.Fatal("closed admission invoked routes")
	}
}

func TestEmptyRoutesReturnCanonical404ForEveryMethod(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	routes, err := NewRoutes(func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	handler := newTestHandler(t, admission, routes, nil)
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch} {
		recorder := requestEdge(handler, method, "/anything")
		assertEdgeFailure(t, recorder, http.StatusNotFound, "ROUTE_NOT_FOUND")
	}
}

func TestEdgeRejectsNonCanonicalPathsWithoutRedirect(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	handler := newTestHandler(t, admission, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), nil)
	for _, path := range []string{"/path/", "/a/../path", "//path", "/back\\slash"} {
		recorder := requestEdge(handler, http.MethodGet, path)
		assertEdgeFailure(t, recorder, http.StatusNotFound, "ROUTE_NOT_FOUND")
	}
}

func TestEdgePropagatesCorrelationContext(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	handler := newTestHandler(t, admission, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, found := requestmeta.CorrelationIDFromContext(r.Context())
		if !found {
			t.Fatal("correlation ID missing from context")
		}
		w.Header().Set("X-Observed-Correlation-ID", id.String())
		w.WriteHeader(http.StatusNoContent)
	}), nil)
	recorder := requestEdge(handler, http.MethodGet, "/resource")
	if recorder.Code != http.StatusNoContent || recorder.Header().Get("X-Observed-Correlation-ID") != "edge-test" {
		t.Fatalf("unexpected correlation propagation: %d %#v", recorder.Code, recorder.Header())
	}
}

func TestEdgePanicDiscardsBufferedResponse(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	var reported []error
	handler := newTestHandler(t, admission, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("must-not-escape"))
		panic("sensitive panic value")
	}), func(err error) { reported = append(reported, err) })
	recorder := requestEdge(handler, http.MethodPost, "/resource")
	assertEdgeFailure(t, recorder, http.StatusInternalServerError, "INTERNAL_ERROR")
	if strings.Contains(recorder.Body.String(), "must-not-escape") || strings.Contains(recorder.Body.String(), "sensitive") {
		t.Fatal("panic response leaked buffered or panic content")
	}
	if len(reported) != 1 || strings.Contains(reported[0].Error(), "sensitive") {
		t.Fatalf("unexpected panic report: %#v", reported)
	}
}

func newTestHandler(t *testing.T, admission *Admission, routes http.Handler, report ErrorSink) *Handler {
	t.Helper()
	if report == nil {
		report = func(error) {}
	}
	handler, err := NewHandler(admission, routes, report, func(context.Context, RequestObservation) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func requestEdge(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set(requestmeta.CorrelationHeader, "edge-test")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func assertEdgeFailure(t *testing.T, recorder *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if recorder.Code != status {
		t.Fatalf("expected status %d, got %d", status, recorder.Code)
	}
	var envelope runtimev1.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != code || envelope.Error.CorrelationID != "edge-test" {
		t.Fatalf("unexpected failure: %#v", envelope)
	}
}
