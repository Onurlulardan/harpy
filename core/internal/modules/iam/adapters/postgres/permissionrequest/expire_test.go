package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExpireRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewPermissionRequestStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Expire(
		context.Background(), platformtx.TenantScope{}, permissionrequest.ExpireCommand{},
	); err == nil {
		t.Fatal("expire accepted an unowned tenant scope")
	}
}
