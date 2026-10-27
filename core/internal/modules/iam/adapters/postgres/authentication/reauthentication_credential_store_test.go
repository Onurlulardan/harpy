package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestReauthenticationCredentialStoreRejectsUnownedScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.ReauthenticationCredential(
		context.Background(), authentication.CredentialScope{},
		authentication.ReauthenticationLookup{},
	)
	if err == nil {
		t.Fatal("credential read accepted an unowned scope")
	}
}
