package iam

import "testing"

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
	if len(routes) != len(want) {
		t.Fatalf("route count=%d want %d", len(routes), len(want))
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
