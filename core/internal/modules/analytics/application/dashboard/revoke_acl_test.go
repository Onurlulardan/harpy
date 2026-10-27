package dashboard

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestRevokeACLRequiresIdempotencyKey(t *testing.T) {
	err := (&Application{}).RevokeACL(context.Background(), platformtx.TenantScope{}, RevokeACLRequest{DashboardID: dashboardFixtureID, PrincipalID: dashboardFixturePrincipal, PrincipalKind: "user", ExpectedRevision: 1})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("revoke ACL error = %v", err)
	}
}
