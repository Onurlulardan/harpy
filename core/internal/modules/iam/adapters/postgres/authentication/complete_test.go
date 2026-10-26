package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestCompleteReauthenticationRejectsUnownedScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CompleteReauthentication(
		context.Background(), authentication.CredentialScope{},
		authentication.ReauthenticationCompletion{},
	)
	if err == nil {
		t.Fatal("reauthentication completion accepted an unowned scope")
	}
}
