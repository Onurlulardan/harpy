package apidocs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerFailsClosedWithoutAnEnabledValidatedBoundary(t *testing.T) {
	disabled, err := New(false, []byte("not JSON"))
	if err != nil {
		t.Fatalf("disabled boundary inspected an unused artifact: %v", err)
	}
	responses := []*httptest.ResponseRecorder{
		request(t, disabled, http.MethodGet, "/swagger/"),
	}
	var nilHandler *Handler
	responses = append(responses, request(t, nilHandler, http.MethodGet, "/swagger/"))

	nilRequest := httptest.NewRecorder()
	disabled.ServeHTTP(nilRequest, nil)
	responses = append(responses, nilRequest)
	nilURL := httptest.NewRecorder()
	disabled.ServeHTTP(nilURL, &http.Request{Method: http.MethodGet})
	responses = append(responses, nilURL)

	for index, response := range responses {
		if response.Code != http.StatusNotFound {
			t.Fatalf("response %d status=%d", index, response.Code)
		}
		assertBoundaryHeaders(t, response)
	}
}

func TestHandlerDistinguishesUnknownPathsFromUnsupportedMethods(t *testing.T) {
	handler, err := New(true, []byte(testDocument))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/swagger/missing"},
		{http.MethodPost, "/swagger/missing"},
		{http.MethodPost, "/swagger/../../secrets"},
		{http.MethodGet, "/swaggerish"},
	} {
		response := request(t, handler, test.method, test.path)
		if response.Code != http.StatusNotFound || response.Header().Get("Allow") != "" {
			t.Fatalf("%s %s status=%d allow=%q", test.method, test.path, response.Code, response.Header().Get("Allow"))
		}
		assertBoundaryHeaders(t, response)
	}

	response := request(t, handler, http.MethodPost, "/swagger/openapi.json")
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("status=%d allow=%q", response.Code, response.Header().Get("Allow"))
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		t.Fatalf("content-type=%q", contentType)
	}
	assertBoundaryHeaders(t, response)
}
