package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewPermissionRequestStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(
		context.Background(), platformtx.TenantScope{}, permissionrequest.CreateCommand{},
	); err == nil {
		t.Fatal("create accepted an unowned tenant scope")
	}
}
