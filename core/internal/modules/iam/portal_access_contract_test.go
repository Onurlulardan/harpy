package iam

import (
	"context"
	"errors"
	"testing"
)

func TestPortalAccessRuntimeFailsClosedWithoutAuthority(t *testing.T) {
	if _, err := (*PortalAccessRuntime)(nil).VerifyPortalAccess(
		context.Background(), PortalAccessRequest{},
	); !errors.Is(err, ErrPortalAccessUnavailable) {
		t.Fatalf("verify error=%v", err)
	}
	if err := (*PortalAccessRuntime)(nil).AuthorizePortalAction(
		context.Background(), PortalAccessDecision{}, "documents.items", "read",
	); !errors.Is(err, ErrPortalAccessUnavailable) {
		t.Fatalf("authorize error=%v", err)
	}
}

func TestPortalAccessAuthenticationHeadersAreExact(t *testing.T) {
	for _, values := range [][]string{nil, {""}, {"Bearer "},
		{"Bearer first", "Bearer second"}, {"Basic token"}} {
		if _, valid := formsRecordBearer(values); valid {
			t.Fatalf("invalid portal bearer accepted: %q", values)
		}
	}
	if token, valid := formsRecordBearer([]string{"Bearer portal-token"}); !valid || token != "portal-token" {
		t.Fatalf("exact portal bearer rejected: %q %t", token, valid)
	}
}
