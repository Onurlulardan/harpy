package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExecuteOwnerTransferCompensationUsesABAGuardRoutine(t *testing.T) {
	if !strings.Contains(applyOwnerTransferCompensationSQL, "compensate_apply_v1") {
		t.Fatal("compensation apply SQL lacks ABA guard routine")
	}
	var storage *store
	if _, err := storage.ExecuteOwnerTransferCompensation(context.Background(),
		platformtx.TenantScope{}, platformdomain.OwnerTransferAcknowledge{}); err == nil {
		t.Fatal("compensation apply accepted unbound scope")
	}
}
