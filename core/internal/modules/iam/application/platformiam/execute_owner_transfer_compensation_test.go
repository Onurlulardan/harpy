package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExecuteOwnerTransferCompensationFailsClosed(t *testing.T) {
	var service *OwnerTransferService
	if _, err := service.ExecuteOwnerTransferCompensation(context.Background(),
		platformtx.TenantScope{}, platformdomain.OwnerTransferAcknowledge{}); err == nil {
		t.Fatal("owner transfer compensation ack accepted an unowned scope")
	}
}
