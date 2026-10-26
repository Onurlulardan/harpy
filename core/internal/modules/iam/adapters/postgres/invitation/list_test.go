package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewInvitationStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.List(
		context.Background(), platformtx.TenantScope{}, invitation.ListQuery{},
	); err == nil {
		t.Fatal("invitation list accepted an unowned tenant scope")
	}
}
