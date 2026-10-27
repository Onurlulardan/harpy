package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAuthorizeFailsClosedOutsideOwnedSnapshotScope(t *testing.T) {
	var service *Service
	if _, err := service.Authorize(context.Background(), platformtx.TenantScope{},
		platformdomain.EvaluateQuery{}); err == nil {
		t.Fatal("authorize accepted an unowned scope")
	}
}
