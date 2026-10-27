package iam

import (
	"net/http"
	"testing"
)

func TestInvitationHTTPRoutesPreserveV1Surface(t *testing.T) {
	routes := (&InvitationHTTP{}).Routes()
	want := map[string]bool{
		http.MethodPost + " /api/invitations":        false,
		http.MethodGet + " /api/invitations":         false,
		http.MethodGet + " /api/invitations/{id}":    false,
		http.MethodDelete + " /api/invitations/{id}": false,
		http.MethodGet + " /api/invitations/verify":  false,
		http.MethodPost + " /api/invitations/accept": false,
	}
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		if _, found := want[key]; !found {
			t.Fatalf("unexpected invitation route %q", key)
		}
		want[key] = true
	}
	for key, found := range want {
		if !found {
			t.Fatalf("missing invitation route %q", key)
		}
	}
}
