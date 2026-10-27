package iam

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLifecycleHTTPRoutesExposeExactV2Surface(t *testing.T) {
	routes := (&LifecycleHTTP{}).Routes()
	want := []string{
		"POST /api/iam/identifiers",
		"POST /api/iam/identifiers/{id}/verification",
		"POST /api/auth/identifiers/verify",
		"POST /api/iam/identifiers/{id}/tombstone",
		"POST /api/iam/identifiers/{id}/retire",
		"POST /api/auth/password/reset",
		"POST /api/auth/password/reset/complete",
		"POST /api/auth/initial-admin/setup/complete",
		"POST /api/iam/password/change",
		"GET /api/iam/portal-setups/{binding_id}",
		"POST /api/iam/portal-setups/{binding_id}/token",
		"POST /api/auth/portal/setup/complete",
		"POST /api/iam/portal-setups/{binding_id}/decision",
		"POST /api/iam/authenticators/webauthn/enrollments",
		"POST /api/iam/authenticators/webauthn/enrollments/{authenticator_id}/complete",
		"POST /api/iam/authenticators/totp/enrollments",
		"POST /api/iam/authenticators/totp/enrollments/{authenticator_id}/complete",
		"POST /api/iam/authenticators/recovery-codes",
		"GET /api/iam/authenticators/recovery-codes",
		"GET /api/iam/authenticators",
		"POST /api/iam/authenticators/{id}/suspend",
		"POST /api/iam/authenticators/{id}/retire",
		"GET /api/iam/sessions",
		"DELETE /api/iam/sessions/{id}",
		"DELETE /api/iam/sessions",
	}
	if len(routes) < len(want) {
		t.Fatalf("route count=%d want at least %d", len(routes), len(want))
	}
	seen := map[string]bool{}
	for index, expected := range want {
		actual := routes[index].Method + " " + routes[index].Pattern
		if actual != expected || seen[actual] {
			t.Fatalf("route[%d]=%q want %q duplicate=%t", index, actual, expected, seen[actual])
		}
		seen[actual] = true
	}
}

func TestLifecycleHTTPRetiredV1RoutesReturnGoneWithoutEcho(t *testing.T) {
	handler := &LifecycleHTTP{report: func(error) {}}
	routes := handler.Routes()[25:]
	if len(routes) != 21 {
		t.Fatalf("retired route count=%d want 21", len(routes))
	}
	secret := "v1-secret-must-not-echo"
	portalUnlinkFound := false
	for _, route := range routes {
		if route.Method == http.MethodPost &&
			route.Pattern == "/api/auth/register-tenant" {
			t.Fatal("tenancy-owned legacy registration route leaked into IAM")
		}
		portalUnlink := route.Method == http.MethodDelete &&
			route.Pattern == "/api/auth/portal-users/{id}/link"
		portalUnlinkFound = portalUnlinkFound || portalUnlink
		request := httptest.NewRequest(route.Method, route.Pattern+"?token="+secret,
			strings.NewReader(`{"password":"`+secret+`"}`))
		response := httptest.NewRecorder()
		route.Handler.ServeHTTP(response, request)
		if response.Code != http.StatusGone ||
			response.Header().Get("Cache-Control") != "no-store" ||
			!bytes.Contains(response.Body.Bytes(), []byte("CLIENT_UPGRADE_REQUIRED")) ||
			bytes.Contains(response.Body.Bytes(), []byte(secret)) {
			t.Fatalf("retired %s %s status=%d headers=%v body=%s",
				route.Method, route.Pattern, response.Code, response.Header(), response.Body.Bytes())
		}
		if portalUnlink && response.Header().Get("Link") != "" {
			t.Fatalf("retired portal-user unlink advertised an invented successor: %q",
				response.Header().Get("Link"))
		}
	}
	if !portalUnlinkFound {
		t.Fatal("retired V1 portal-user unlink route is missing")
	}
}
