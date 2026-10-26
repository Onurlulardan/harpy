package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestRotatePostgresRejectsUnownedCredentialScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Rotate(
		context.Background(), authentication.CredentialScope{},
		authentication.RefreshRotation{},
	); err == nil {
		t.Fatal("refresh rotation accepted an unowned credential scope")
	}
}
