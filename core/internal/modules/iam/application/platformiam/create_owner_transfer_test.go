package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateOwnerTransferFailsClosedOutsideOwnedScope(t *testing.T) {
	var service *OwnerTransferService
	if _, err := service.CreateOwnerTransfer(context.Background(), platformtx.TenantScope{},
		platformdomain.OwnerTransferPrepare{}); err == nil {
		t.Fatal("owner transfer prepare accepted an unowned scope")
	}
}
