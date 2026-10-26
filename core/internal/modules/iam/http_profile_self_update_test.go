package iam

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestProfileHTTPSelfUpdateRejectsUnknownFieldWithStable422(t *testing.T) {
	fixture := newProfileHTTPTestFixture(t)
	request := fixture.request(http.MethodPut, "/api/profile", `{"is_admin":true}`)
	response := fixture.serve(request)
	if response.Code != http.StatusUnprocessableEntity || len(fixture.app.calls) != 0 ||
		!strings.Contains(response.Body.String(), `"code":"PROFILE_FIELD_FORBIDDEN"`) {
		t.Fatalf("status=%d calls=%d body=%s", response.Code, len(fixture.app.calls), response.Body)
	}
}

func TestProfileHTTPSelfUpdatePreservesNullDeletion(t *testing.T) {
	fixture := newProfileHTTPTestFixture(t)
	fixture.app.results["UpdateSelf"] = profileHTTPTestSelfView()
	request := fixture.request(http.MethodPut, "/api/profile", `{"avatar":null}`)
	response := fixture.serve(request)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code, len(fixture.app.calls), response.Body)
	}
	input, ok := fixture.app.calls[0].input.(ProfileSelfUpdate)
	if !ok || string(input.Values["avatar"]) != "null" || input.IdempotencyKey == "" {
		t.Fatalf("self update input=%+v", fixture.app.calls[0].input)
	}
}

func TestProfileHTTPSelfPresenterFailsClosedOnSecret(t *testing.T) {
	fixture := newProfileHTTPTestFixture(t)
	fixture.app.results["GetSelf"] = ProfileSelfView{
		Document: json.RawMessage(`{"id":"51000000-0000-4000-8000-000000000001",` +
			`"password_hash":"secret"}`),
		Revision: 2,
	}
	response := fixture.serve(fixture.request(http.MethodGet, "/api/profile", ""))
	if response.Code != http.StatusServiceUnavailable ||
		strings.Contains(response.Body.String(), "secret") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}
