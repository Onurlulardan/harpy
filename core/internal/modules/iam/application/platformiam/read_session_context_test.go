package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReadSessionContextFailsClosedOutsideOwnedScope(t *testing.T) {
	var service *Service
	if _, err := service.ReadSessionContext(context.Background(),
		platformtx.TenantScope{}, platformdomain.SessionContextQuery{}); err == nil {
		t.Fatal("session context accepted an unowned scope")
	}
}
