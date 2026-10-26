package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCompleteOwnerTransferFailsClosedOutsideOwnedScope(t *testing.T) {
	var service *OwnerTransferService
	if _, err := service.CompleteOwnerTransfer(context.Background(), platformtx.TenantScope{},
		platformdomain.OwnerTransferCommit{}); err == nil {
		t.Fatal("owner transfer commit accepted an unowned scope")
	}
}
