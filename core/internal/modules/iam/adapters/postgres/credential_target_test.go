package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCredentialTargetBinderAndSQLExposeOnlyNarrowResolvers(t *testing.T) {
	if _, err := NewCredentialTargetResolverBinder(); err != nil {
		t.Fatalf("credential target binder: %v", err)
	}
	tests := []struct {
		document string
		routine  string
	}{
		{resolveLoginCredentialTargetSQL, "platform.iam_resolve_login_target_v7"},
		{resolveRefreshCredentialTargetSQL, "platform.iam_resolve_refresh_target_v2"},
		{resolveTenantCredentialTargetSQL, "platform.iam_resolve_tenant_target_v1"},
	}
	for _, test := range tests {
		if !strings.Contains(test.document, test.routine) ||
			strings.Contains(test.document, "platform.iam_login_directory") ||
			strings.Contains(test.document, "platform.iam_refresh_locators") ||
			strings.Contains(test.document, "platform.domain_cutovers") {
			t.Fatalf("credential resolver SQL is not narrow: %s", test.document)
		}
	}
}

func TestCredentialTargetResolverFailsClosedWithoutBoundDatabase(t *testing.T) {
	resolver := &credentialTargetResolver{}
	lookup, err := platformtx.NewRefreshCredentialLookup(
		"iam.refresh", [16]byte{1},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.CredentialTarget(
		context.Background(), lookup,
	); !errors.Is(err, errCredentialTargetUnavailable) {
		t.Fatalf("unbound resolver error = %v", err)
	}
	if _, err := (*credentialTargetResolver)(nil).CredentialTarget(
		context.Background(), lookup,
	); !errors.Is(err, errCredentialTargetUnavailable) {
		t.Fatalf("nil resolver error = %v", err)
	}
}
