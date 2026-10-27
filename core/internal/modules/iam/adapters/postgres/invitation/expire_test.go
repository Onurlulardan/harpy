package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExpireRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewInvitationStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Expire(
		context.Background(), platformtx.TenantScope{}, invitation.ExpireCommand{},
	); err == nil {
		t.Fatal("invitation expire accepted an unowned tenant scope")
	}
}
