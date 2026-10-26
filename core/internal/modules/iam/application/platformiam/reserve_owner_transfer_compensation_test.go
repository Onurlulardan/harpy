package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReserveOwnerTransferCompensationFailsClosed(t *testing.T) {
	var service *OwnerTransferService
	if _, err := service.ReserveOwnerTransferCompensation(context.Background(),
		platformtx.TenantScope{}, platformdomain.OwnerTransferCompensation{}); err == nil {
		t.Fatal("owner transfer compensation accepted an unowned scope")
	}
}
