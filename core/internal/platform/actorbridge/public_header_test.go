package actorbridge

import (
	"net/http"
	"net/http/httptest"
	"testing"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func TestPublicGuardRejectsActorAssertionHeaderCaseInsensitively(t *testing.T) {
	called := 0
	guard, err := RejectPublicActorAssertion(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			called++
			writer.WriteHeader(http.StatusNoContent)
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	clean := httptest.NewRequest(http.MethodGet, "https://api.invalid/health", nil)
	cleanResponse := httptest.NewRecorder()
	guard.ServeHTTP(cleanResponse, clean)
	if cleanResponse.Code != http.StatusNoContent || called != 1 {
		t.Fatalf("clean request status=%d called=%d", cleanResponse.Code, called)
	}
	for _, name := range []string{
		actorassertionv1.HeaderName, "harpy-actor-assertion",
		"HARPY-ACTOR-ASSERTION",
	} {
		request := httptest.NewRequest(http.MethodGet, "https://api.invalid/", nil)
		request.Header = http.Header{name: []string{"value"}}
		response := httptest.NewRecorder()
		guard.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || called != 1 {
			t.Fatalf("header %s status=%d called=%d", name, response.Code, called)
		}
	}
}
