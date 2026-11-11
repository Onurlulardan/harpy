package apidocs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIReferenceIsServedBesideThePublicAPI(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" {
			t.Errorf("public path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler, err := Mount(next)
	if err != nil {
		t.Fatal(err)
	}
	page := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8000/docs", nil)
	pageResponse := httptest.NewRecorder()
	handler.ServeHTTP(pageResponse, page)
	body := pageResponse.Body.String()
	if pageResponse.Code != http.StatusOK ||
		!strings.Contains(pageResponse.Header().Get("Content-Type"), "text/html") ||
		!strings.Contains(body, "Scalar.createApiReference") ||
		!strings.Contains(body, `"title":"Core API"`) ||
		!strings.Contains(body, `"title":"Management"`) ||
		!strings.Contains(body, `"persistAuth":true`) ||
		!strings.Contains(body, `"hideModels":true`) ||
		!strings.Contains(body, "harpy-docs-access-token") ||
		!strings.Contains(body, "/api/auth/login") ||
		!strings.Contains(body, "/api/auth/initial-password") ||
		!strings.Contains(body, "http://127.0.0.1:8000/docs/spec/http/core-api-v1.openapi.yaml") ||
		!strings.Contains(body, "http://127.0.0.1:8000/docs/spec/http/management-v1.openapi.yaml") {
		t.Fatalf("reference status=%d body=%s", pageResponse.Code, body)
	}

	spec := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8000/docs/spec/http/core-api-v1.openapi.yaml", nil)
	specResponse := httptest.NewRecorder()
	handler.ServeHTTP(specResponse, spec)
	if specResponse.Code != http.StatusOK || !strings.Contains(specResponse.Body.String(), `"openapi": "3.1.1"`) {
		t.Fatalf("core spec status=%d", specResponse.Code)
	}
	schema := httptest.NewRequest(http.MethodGet,
		"http://127.0.0.1:8000/docs/spec/runtime/v1/error-envelope.schema.json", nil)
	schemaResponse := httptest.NewRecorder()
	handler.ServeHTTP(schemaResponse, schema)
	if schemaResponse.Code != http.StatusOK || !strings.Contains(schemaResponse.Body.String(), "correlation_id") {
		t.Fatalf("error schema status=%d", schemaResponse.Code)
	}

	api := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8000/api/auth/login", nil)
	apiResponse := httptest.NewRecorder()
	handler.ServeHTTP(apiResponse, api)
	if apiResponse.Code != http.StatusNoContent {
		t.Fatalf("public API status=%d", apiResponse.Code)
	}
}

func TestAPIReferenceRejectsSpecTraversal(t *testing.T) {
	if specName("http/../files.go") || specName("runtime/v1/../../files.go") {
		t.Fatal("spec name accepted a parent segment")
	}
	handler, err := Mount(http.NotFoundHandler())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8000/docs/spec/files.go", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("embedded source status=%d", response.Code)
	}
}
