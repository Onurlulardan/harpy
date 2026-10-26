package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestManagementRejectsUnknownAndNonCanonicalPathsWithoutRedirect(t *testing.T) {
	fixture := newHandlerFixture(t)
	for _, path := range []string{"/", "/livez/", "/a/../livez", "//livez"} {
		recorder := requestProbe(t, fixture.handler, http.MethodGet, path)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for %s, got %d", path, recorder.Code)
		}
		assertFailureCode(t, recorder, "ROUTE_NOT_FOUND")
	}
}

func TestManagementRejectsWrongMethodWithAllowHeader(t *testing.T) {
	fixture := newHandlerFixture(t)
	recorder := requestProbe(t, fixture.handler, http.MethodPost, "/livez")
	if recorder.Code != http.StatusMethodNotAllowed || recorder.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("unexpected method response: %d %#v", recorder.Code, recorder.Header())
	}
	assertFailureCode(t, recorder, "METHOD_NOT_ALLOWED")
}

func TestManagementReplacesDuplicateCorrelationID(t *testing.T) {
	fixture := newHandlerFixture(t)
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	request.Header.Add("X-Correlation-ID", "first")
	request.Header.Add("X-Correlation-ID", "second")
	recorder := httptest.NewRecorder()
	fixture.handler.ServeHTTP(recorder, request)
	correlationID := recorder.Header().Get("X-Correlation-ID")
	if correlationID == "first" || correlationID == "second" || len(correlationID) != 36 {
		t.Fatalf("duplicate correlation ID was not replaced: %q", correlationID)
	}
	var envelope runtimev1.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.CorrelationID != correlationID {
		t.Fatalf("header/body correlation mismatch")
	}
}

func TestManagementDoesNotEmitCORSHeaders(t *testing.T) {
	fixture := newHandlerFixture(t)
	recorder := requestProbe(t, fixture.handler, http.MethodGet, "/livez")
	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("management endpoint emitted a CORS header")
	}
}

func assertFailureCode(t *testing.T, recorder *httptest.ResponseRecorder, expected string) {
	t.Helper()
	var envelope runtimev1.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != expected || envelope.Error.Issues == nil {
		t.Fatalf("unexpected failure envelope: %#v", envelope)
	}
}
