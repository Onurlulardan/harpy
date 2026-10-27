package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewPermissionRequestStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.List(
		context.Background(), platformtx.TenantScope{}, permissionrequest.ListQuery{},
	); err == nil {
		t.Fatal("list accepted an unowned tenant scope")
	}
}
