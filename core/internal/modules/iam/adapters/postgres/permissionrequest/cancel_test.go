package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCancelRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewPermissionRequestStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Cancel(
		context.Background(), platformtx.TenantScope{}, permissionrequest.CancelCommand{},
	); err == nil {
		t.Fatal("cancel accepted an unowned tenant scope")
	}
}
