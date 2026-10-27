package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRevokeFailsClosedOutsideOwnedCommandScope(t *testing.T) {
	var service *Service
	if _, err := service.Revoke(context.Background(), platformtx.TenantScope{},
		platformdomain.OperatorRevoke{}); err == nil {
		t.Fatal("revoke accepted an unowned scope")
	}
}
