package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestRevokeSessionRejectsUnownedScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.RevokeSession(
		context.Background(), authentication.CredentialScope{},
		authentication.LogoutMutation{},
	)
	if err == nil {
		t.Fatal("session revocation accepted an unowned scope")
	}
}
