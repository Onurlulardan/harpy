package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewInvitationStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(
		context.Background(), platformtx.TenantScope{}, invitation.CreateCommand{},
	); err == nil {
		t.Fatal("invitation create accepted an unowned tenant scope")
	}
}
