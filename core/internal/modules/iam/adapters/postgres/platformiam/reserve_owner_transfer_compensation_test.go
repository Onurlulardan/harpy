package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReserveOwnerTransferCompensationUsesCASRoutine(t *testing.T) {
	if !strings.Contains(prepareOwnerTransferCompensationSQL, "compensate_prepare_v1") {
		t.Fatal("compensation prepare SQL lacks CAS routine")
	}
	var storage *store
	if _, err := storage.ReserveOwnerTransferCompensation(context.Background(),
		platformtx.TenantScope{}, platformdomain.OwnerTransferCompensation{}); err == nil {
		t.Fatal("compensation prepare accepted unbound scope")
	}
}
