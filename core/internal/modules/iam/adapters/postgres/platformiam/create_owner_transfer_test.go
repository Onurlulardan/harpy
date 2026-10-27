package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateOwnerTransferUsesNarrowRoutine(t *testing.T) {
	for _, marker := range []string{"iam_owner_transfer_prepare_v1", "$28::timestamptz"} {
		if !strings.Contains(prepareOwnerTransferSQL, marker) {
			t.Fatalf("prepare SQL lacks %q", marker)
		}
	}
	var storage *store
	if _, err := storage.CreateOwnerTransfer(context.Background(), platformtx.TenantScope{},
		platformdomain.OwnerTransferPrepare{}); err == nil {
		t.Fatal("prepare accepted unbound scope")
	}
}
