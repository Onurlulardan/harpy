package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestIssuePostgresRejectsUnownedCredentialScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Issue(
		context.Background(), authentication.CredentialScope{},
		authentication.SessionIssue{},
	); err == nil {
		t.Fatal("session issue accepted an unowned credential scope")
	}
}
