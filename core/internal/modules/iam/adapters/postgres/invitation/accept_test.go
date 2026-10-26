package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAcceptRejectsUnownedCredentialScope(t *testing.T) {
	store, err := NewInvitationStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Accept(
		context.Background(), platformtx.CredentialTenantScope{},
		invitation.AcceptCommand{},
	); err == nil {
		t.Fatal("invitation accept accepted an unowned credential scope")
	}
}
