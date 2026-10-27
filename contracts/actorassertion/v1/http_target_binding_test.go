package actorassertionv1

import (
	"crypto/sha256"
	"strings"
	"testing"
)

func TestHTTPRequestTargetBindingIsDeterministicAndRequestBound(t *testing.T) {
	body := sha256.Sum256([]byte(`{"name":"alpha"}`))
	first, err := HTTPRequestTargetBindingSHA256(
		"POST", "/api/forms/records", "page=1", body, 16,
	)
	if err != nil {
		t.Fatalf("build HTTP target binding: %v", err)
	}
	replay, err := HTTPRequestTargetBindingSHA256(
		"POST", "/api/forms/records", "page=1", body, 16,
	)
	if err != nil || replay != first {
		t.Fatalf("HTTP target binding replay = %x, %v", replay, err)
	}
	changed, err := HTTPRequestTargetBindingSHA256(
		"POST", "/api/forms/records", "page=2", body, 16,
	)
	if err != nil || changed == first {
		t.Fatal("HTTP target binding ignored the raw query")
	}
}

func TestHTTPRequestTargetBindingRejectsOutsideClosedProfile(t *testing.T) {
	body := sha256.Sum256(nil)
	tests := []struct {
		method, path, query string
		length              int64
	}{
		{"OPTIONS", "/api/forms", "", 0},
		{"GET", "api/forms", "", 0},
		{"GET", "/" + strings.Repeat("a", MaximumHTTPBindingPathBytes), "", 0},
		{"GET", "/api/forms", strings.Repeat("q", MaximumHTTPBindingQueryBytes+1), 0},
		{"POST", "/api/forms", "", -1},
		{"POST", "/api/forms", "", MaximumHTTPBindingBodyBytes + 1},
	}
	for _, test := range tests {
		if _, err := HTTPRequestTargetBindingSHA256(
			test.method, test.path, test.query, body, test.length,
		); err == nil {
			t.Fatalf("accepted invalid HTTP binding: %+v", test)
		}
	}
}
