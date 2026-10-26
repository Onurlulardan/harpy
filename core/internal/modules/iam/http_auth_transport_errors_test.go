package iam

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthHTTPLoginPublishesStableTransportFailures(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	body := `{"tenant_slug":"tenant-a","email":"user@example.com","password":"secret"}`
	tests := []struct {
		name          string
		contentType   string
		contentLength int64
		wantStatus    int
		wantCode      string
	}{
		{"media", "", int64(len(body)), http.StatusUnsupportedMediaType, "MEDIA_TYPE_UNSUPPORTED"},
		{"size", "application/json", authLoginBodyLimit + 1, http.StatusRequestEntityTooLarge, "REQUEST_BODY_TOO_LARGE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
			request.Header.Set("Origin", fixture.origin)
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}
			request.ContentLength = test.contentLength
			response := httptest.NewRecorder()
			fixture.handler.login(response, request)
			document := decodeAuthResponse(t, response)
			if response.Code != test.wantStatus ||
				document["error"].(map[string]any)["code"] != test.wantCode {
				t.Fatalf("transport response=%d %v", response.Code, document)
			}
		})
	}
	if fixture.app.loginCalls != 0 {
		t.Fatalf("invalid transport reached application %d times", fixture.app.loginCalls)
	}
}

func TestAuthHTTPApplicationFailureIsRedactedAndUnavailable(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	fixture.app.loginError = errors.New("database password secret-value failed")
	response := executeRejectedLogin(t, fixture, "tenant-a")
	if response.Code != http.StatusServiceUnavailable ||
		strings.Contains(response.Body.String(), "database") ||
		strings.Contains(response.Body.String(), "secret-value") ||
		len(fixture.errors) != 1 {
		t.Fatalf("unredacted unavailable failure: %d %s %v", response.Code, response.Body.String(), fixture.errors)
	}
}
