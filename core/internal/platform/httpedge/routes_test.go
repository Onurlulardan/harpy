package httpedge

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutesDispatchExactMethodAndPathParameters(t *testing.T) {
	routes, err := NewRoutes(func(error) {},
		Route{Method: http.MethodGet, Pattern: "/api/users/{user_id}", Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-User-ID", r.PathValue("user_id"))
				w.WriteHeader(http.StatusNoContent)
			},
		)},
		Route{Method: http.MethodGet, Pattern: "/api/users/current", Handler: http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Static", "true")
				w.WriteHeader(http.StatusNoContent)
			},
		)},
	)
	if err != nil {
		t.Fatal(err)
	}
	dynamic := httptest.NewRecorder()
	routes.ServeHTTP(dynamic, httptest.NewRequest(http.MethodGet, "/api/users/123", nil))
	if dynamic.Code != http.StatusNoContent || dynamic.Header().Get("X-User-ID") != "123" {
		t.Fatalf("dynamic route mismatch: %d %v", dynamic.Code, dynamic.Header())
	}
	static := httptest.NewRecorder()
	routes.ServeHTTP(static, httptest.NewRequest(http.MethodGet, "/api/users/current", nil))
	if static.Code != http.StatusNoContent || static.Header().Get("X-Static") != "true" ||
		static.Header().Get("X-User-ID") != "" {
		t.Fatalf("static precedence mismatch: %d %v", static.Code, static.Header())
	}
}

func TestRoutesReturnCanonicalNotFoundAndMethodFailure(t *testing.T) {
	routes, err := NewRoutes(func(error) {},
		Route{Method: http.MethodGet, Pattern: "/resource", Handler: http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) },
		)},
		Route{Method: http.MethodDelete, Pattern: "/resource", Handler: http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) },
		)},
	)
	if err != nil {
		t.Fatal(err)
	}
	notFound := httptest.NewRecorder()
	routes.ServeHTTP(notFound, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d", notFound.Code)
	}
	method := httptest.NewRecorder()
	routes.ServeHTTP(method, httptest.NewRequest(http.MethodPost, "/resource", nil))
	if method.Code != http.StatusMethodNotAllowed || method.Header().Get("Allow") != "DELETE, GET" {
		t.Fatalf("method response = %d %v", method.Code, method.Header())
	}
}

func TestRoutesRejectInvalidDuplicateAndAmbiguousContracts(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	tests := [][]Route{
		{{Method: "get", Pattern: "/resource", Handler: handler}},
		{{Method: http.MethodGet, Pattern: "/resource", Handler: handler, Precedence: -1}},
		{{Method: http.MethodGet, Pattern: "/resource", Handler: handler, Precedence: 101}},
		{{Method: http.MethodGet, Pattern: "/Resource", Handler: handler}},
		{{Method: http.MethodGet, Pattern: "/resource/", Handler: handler}},
		{{Method: http.MethodGet, Pattern: "/{bad-name}", Handler: handler}},
		{
			{Method: http.MethodGet, Pattern: "/resource", Handler: handler},
			{Method: http.MethodGet, Pattern: "/resource", Handler: handler},
		},
		{
			{Method: http.MethodGet, Pattern: "/{kind}/fixed", Handler: handler},
			{Method: http.MethodPost, Pattern: "/fixed/{id}", Handler: handler},
		},
		{
			{Method: http.MethodGet, Pattern: "/resource/{id}", Handler: handler},
			{Method: http.MethodPost, Pattern: "/resource/{resource_id}", Handler: handler},
		},
	}
	for index, entries := range tests {
		if _, err := NewRoutes(func(error) {}, entries...); err == nil {
			t.Fatalf("invalid route set %d was accepted", index)
		}
	}
	if _, err := NewRoutes(nil); err == nil {
		t.Fatal("nil error sink was accepted")
	}
}

func TestAmbiguousRouteErrorIdentifiesBothContracts(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	_, err := NewRoutes(func(error) {},
		Route{Method: http.MethodGet, Pattern: "/{kind}/fixed", Handler: handler},
		Route{Method: http.MethodPost, Pattern: "/fixed/{id}", Handler: handler},
	)
	if err == nil || !strings.Contains(err.Error(), "GET /{kind}/fixed") ||
		!strings.Contains(err.Error(), "POST /fixed/{id}") {
		t.Fatalf("ambiguous route error = %v", err)
	}
}

func TestExplicitPrecedenceResolvesIntentionalOverlap(t *testing.T) {
	handler := func(value string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("X-Route", value)
			w.WriteHeader(http.StatusNoContent)
		}
	}
	routes, err := NewRoutes(func(error) {},
		Route{Method: http.MethodGet, Pattern: "/{kind}/fixed", Handler: handler("fallback")},
		Route{Method: http.MethodGet, Pattern: "/fixed/{id}", Handler: handler("explicit"), Precedence: 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	routes.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/fixed/fixed", nil))
	if response.Code != http.StatusNoContent || response.Header().Get("X-Route") != "explicit" {
		t.Fatalf("precedence dispatch = %d %q", response.Code, response.Header().Get("X-Route"))
	}
}
