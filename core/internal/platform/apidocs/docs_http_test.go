package apidocs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testDocument = `{"openapi":"3.1.1","info":{"title":"Test API","version":"v1"},"paths":{"/api/test":{"get":{"operationId":"getTest"}}}}`

func TestHandlerServesDocumentationAndConditionalArtifact(t *testing.T) {
	handler, err := New(true, []byte(testDocument))
	if err != nil {
		t.Fatal(err)
	}
	index := request(t, handler, http.MethodGet, "/swagger/")
	if index.Code != http.StatusOK || !strings.Contains(index.Body.String(), "API documentation") ||
		index.Header().Get("Content-Security-Policy") == "" {
		t.Fatalf("index status=%d body=%s", index.Code, index.Body.String())
	}
	spec := request(t, handler, http.MethodGet, "/swagger/openapi.json")
	if spec.Code != http.StatusOK || spec.Body.String() != testDocument || spec.Header().Get("ETag") == "" {
		t.Fatalf("spec status=%d body=%s etag=%q", spec.Code, spec.Body.String(), spec.Header().Get("ETag"))
	}
	replayRequest := httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil)
	replayRequest.Header.Set("If-None-Match", spec.Header().Get("ETag"))
	replay := httptest.NewRecorder()
	handler.ServeHTTP(replay, replayRequest)
	if replay.Code != http.StatusNotModified || replay.Body.Len() != 0 {
		t.Fatalf("conditional status=%d body=%s", replay.Code, replay.Body.String())
	}
}

func TestHandlerFailsClosedForDisabledInvalidAndUnknownRequests(t *testing.T) {
	disabled, err := New(false, nil)
	if err != nil || request(t, disabled, http.MethodGet, "/swagger/").Code != http.StatusNotFound {
		t.Fatal("disabled documentation did not remain hidden")
	}
	if _, err = New(true, []byte(`{"openapi":"3.0.0"}`)); err == nil {
		t.Fatal("invalid OpenAPI artifact passed")
	}
	enabled, _ := New(true, []byte(testDocument))
	if request(t, enabled, http.MethodPost, "/swagger/").Code != http.StatusMethodNotAllowed ||
		request(t, enabled, http.MethodGet, "/swagger/../../secrets").Code != http.StatusNotFound {
		t.Fatal("unsafe documentation request passed")
	}
}

func request(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
	return response
}
