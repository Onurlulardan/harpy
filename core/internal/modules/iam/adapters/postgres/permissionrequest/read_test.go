package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReadRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewPermissionRequestStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Read(
		context.Background(), platformtx.TenantScope{}, permissionrequest.ReadQuery{},
	); err == nil {
		t.Fatal("read accepted an unowned tenant scope")
	}
}
