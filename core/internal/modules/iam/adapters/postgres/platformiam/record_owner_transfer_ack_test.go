package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRecordOwnerTransferAckUsesNarrowRoutine(t *testing.T) {
	if !strings.Contains(acknowledgeOwnerTransferSQL, "iam_owner_transfer_ack_v1") {
		t.Fatal("ack SQL lacks owner transfer routine")
	}
	var storage *store
	if _, err := storage.RecordOwnerTransferAck(context.Background(), platformtx.TenantScope{},
		platformdomain.OwnerTransferAcknowledge{}); err == nil {
		t.Fatal("ack accepted unbound scope")
	}
}
