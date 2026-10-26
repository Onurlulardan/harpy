package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestRecoveryCredentialStoreRejectsUnownedScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.RecoveryCredential(
		context.Background(), authentication.CredentialScope{},
		authentication.RecoveryLookup{},
	)
	if err == nil {
		t.Fatal("recovery credential resolution accepted an unowned scope")
	}
}
