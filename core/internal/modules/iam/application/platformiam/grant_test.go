package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGrantFailsClosedOutsideOwnedCommandScope(t *testing.T) {
	var service *Service
	if _, err := service.Grant(context.Background(), platformtx.TenantScope{},
		platformdomain.OperatorGrant{}); err == nil {
		t.Fatal("grant accepted an unowned scope")
	}
}
