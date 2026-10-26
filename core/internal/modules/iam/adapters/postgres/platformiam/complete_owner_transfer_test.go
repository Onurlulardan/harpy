package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCompleteOwnerTransferUsesAtomicRoutine(t *testing.T) {
	if !strings.Contains(commitOwnerTransferSQL, "iam_owner_transfer_commit_v1") {
		t.Fatal("commit SQL lacks owner transfer routine")
	}
	var storage *store
	if _, err := storage.CompleteOwnerTransfer(context.Background(), platformtx.TenantScope{},
		platformdomain.OwnerTransferCommit{}); err == nil {
		t.Fatal("commit accepted unbound scope")
	}
}
