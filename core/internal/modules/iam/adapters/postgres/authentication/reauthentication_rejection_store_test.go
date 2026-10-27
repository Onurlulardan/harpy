package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestReauthenticationRejectionStoreRejectsUnownedScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.ReauthenticationRejection(
		context.Background(), authentication.CredentialScope{},
		authentication.ReauthenticationRejection{},
	)
	if err == nil {
		t.Fatal("challenge rejection accepted an unowned scope")
	}
}
