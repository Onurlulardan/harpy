package iam

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIdentityHTTPApplicationErrorsMapToStableRedactedEnvelopes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		status     int
		code       string
		wantReport bool
	}{
		{"invalid", ErrIdentityHTTPInvalid, 400, "REQUEST_INVALID", false},
		{"forbidden", ErrIdentityHTTPForbidden, 403, "AUTHORIZATION_DENIED", false},
		{"not found", ErrIdentityHTTPNotFound, 404, "RESOURCE_NOT_FOUND", false},
		{"conflict", ErrIdentityHTTPConflict, 409, "RESOURCE_CONFLICT", false},
		{"unavailable", ErrIdentityHTTPUnavailable, 503, "IAM_UNAVAILABLE", true},
		{"unknown", errors.New("database secret-value failed"), 503, "IAM_UNAVAILABLE", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newIdentityHTTPTestFixture(t)
			request := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			response := httptest.NewRecorder()
			fixture.handler.writeIdentityApplicationFailure(response, request, test.err)
			if response.Code != test.status ||
				!responseHasIdentityErrorCode(response.Body.Bytes(), test.code) ||
				(len(fixture.errors) != 0) != test.wantReport {
				t.Fatalf("status=%d errors=%v body=%s", response.Code,
					fixture.errors, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "secret-value") ||
				strings.Contains(response.Body.String(), "database") {
				t.Fatal("internal error detail escaped into response")
			}
		})
	}
}

func responseHasIdentityErrorCode(document []byte, want string) bool {
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	return json.Unmarshal(document, &envelope) == nil && envelope.Error.Code == want
}
