package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestConsumeRecoveryRejectsUnownedScope(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.ConsumeRecovery(
		context.Background(), authentication.CredentialScope{},
		authentication.RecoveryConsumption{},
	)
	if err == nil {
		t.Fatal("recovery consumption accepted an unowned scope")
	}
}
