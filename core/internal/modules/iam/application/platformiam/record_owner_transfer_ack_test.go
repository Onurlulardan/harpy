package platformiam

import (
	"context"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRecordOwnerTransferAckFailsClosedOutsideOwnedScope(t *testing.T) {
	var service *OwnerTransferService
	if _, err := service.RecordOwnerTransferAck(context.Background(), platformtx.TenantScope{},
		platformdomain.OwnerTransferAcknowledge{}); err == nil {
		t.Fatal("owner transfer ack accepted an unowned scope")
	}
}
