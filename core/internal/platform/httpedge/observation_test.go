package httpedge

import (
	"context"
	"net/http"
	"testing"
)

func TestEdgeObservesCanonicalRouteAfterResponse(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	routes, err := NewRoutes(func(error) {}, Route{Method: http.MethodGet,
		Pattern: "/api/users/{user_id}", Handler: http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })})
	if err != nil {
		t.Fatal(err)
	}
	var captured RequestObservation
	handler, err := NewHandler(admission, routes, func(error) {},
		func(_ context.Context, value RequestObservation) error {
			captured = value
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	response := requestEdge(handler, http.MethodGet, "/api/users/private-user-id")
	if response.Code != http.StatusCreated || captured.RoutePattern != "/api/users/{user_id}" ||
		captured.Method != http.MethodGet || captured.StatusCode != http.StatusCreated ||
		captured.CorrelationID != "edge-test" || captured.EndedAt.Before(captured.StartedAt) {
		t.Fatalf("unexpected observation: %#v", captured)
	}
}

func TestEdgeDoesNotObserveUnmatchedRawPath(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	routes, err := NewRoutes(func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	handler, err := NewHandler(admission, routes, func(error) {},
		func(context.Context, RequestObservation) error { called = true; return nil })
	if err != nil {
		t.Fatal(err)
	}
	response := requestEdge(handler, http.MethodGet, "/private/raw/identifier")
	if response.Code != http.StatusNotFound || called {
		t.Fatalf("unmatched request observation = %d, %v", response.Code, called)
	}
}
