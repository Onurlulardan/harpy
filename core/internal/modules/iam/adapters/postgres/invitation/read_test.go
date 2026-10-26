package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReadRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewInvitationStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Read(
		context.Background(), platformtx.TenantScope{}, invitation.ReadQuery{},
	); err == nil {
		t.Fatal("invitation read accepted an unowned tenant scope")
	}
}
