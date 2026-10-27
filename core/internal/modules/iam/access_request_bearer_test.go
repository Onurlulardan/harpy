package iam

import "testing"

func TestAccessRequestBearerRejectsMalformedHeaders(t *testing.T) {
	for _, values := range [][]string{
		nil,
		{""},
		{"bearer token"},
		{"Bearer "},
		{"Bearer token with-space"},
		{"Bearer first", "Bearer second"},
	} {
		if _, valid := formsRecordBearer(values); valid {
			t.Fatalf("invalid authorization header accepted: %q", values)
		}
	}
	token, valid := formsRecordBearer([]string{"Bearer exact-token"})
	if !valid || token != "exact-token" {
		t.Fatalf("valid bearer rejected: %q %t", token, valid)
	}
}
