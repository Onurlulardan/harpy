package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestAuthHTTPRoutesPublishExactPhaseFourteenSurface(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	routes := fixture.handler.Routes()
	if len(routes) != 10 {
		t.Fatalf("auth route count=%d", len(routes))
	}
	want := map[string]string{
		"POST /api/auth/login":                     "",
		"POST /api/auth/initial-password":          "",
		"POST /api/auth/portal/login":              "",
		"POST /api/auth/refresh":                   "",
		"POST /api/auth/logout":                    "",
		"GET /api/auth/csrf":                       "",
		"POST /api/auth/csrf":                      "",
		"GET /api/auth/session":                    "",
		"POST /api/auth/reauthentication/start":    "",
		"POST /api/auth/reauthentication/complete": "",
	}
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		if _, found := want[key]; !found || route.Handler == nil {
			t.Fatalf("unexpected auth route: %s", key)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing auth routes: %v", want)
	}
	compiled, err := httpedge.NewRoutes(func(error) {}, routes...)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	response := httptest.NewRecorder()
	compiled.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed ||
		response.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("exact route method response=%d %v", response.Code, response.Header())
	}
}

func TestAuthHTTPConstructorRejectsIncompleteDependencies(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	if _, err := newAuthHTTP(
		fixture.security, nil, func(error) {}, fixture.handler.clock,
	); err == nil {
		t.Fatal("nil application was accepted")
	}
	var typedNil *authHTTPTestApplication
	if _, err := newAuthHTTP(
		fixture.security, typedNil, func(error) {}, fixture.handler.clock,
	); err == nil {
		t.Fatal("typed-nil application was accepted")
	}
}
