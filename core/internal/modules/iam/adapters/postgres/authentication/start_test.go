package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestStartReauthenticationRejectsUnownedScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.StartReauthentication(
		context.Background(), authentication.CredentialScope{},
		authentication.ChallengeCreation{},
	)
	if err == nil {
		t.Fatal("challenge creation accepted an unowned credential scope")
	}
}
