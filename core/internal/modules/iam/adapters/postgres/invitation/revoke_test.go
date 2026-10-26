package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRevokeRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewInvitationStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Revoke(
		context.Background(), platformtx.TenantScope{}, invitation.RevokeCommand{},
	); err == nil {
		t.Fatal("invitation revoke accepted an unowned tenant scope")
	}
}
