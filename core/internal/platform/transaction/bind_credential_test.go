package transaction

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func TestBindCredentialRejectsClosedAndNonIAMBinders(t *testing.T) {
	iamBinder, err := postgres.NewOwnerBinder("iam", func(postgres.BoundExecutor) int { return 1 })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindCredential(CredentialTenantScope{}, iamBinder); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("closed scope error = %v", err)
	}
	foreignBinder, err := postgres.NewOwnerBinder(
		"forms", func(postgres.BoundExecutor) int { return 1 },
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindCredential(
		activeCredentialScopeForTest(t), foreignBinder,
	); !errors.Is(err, postgres.ErrOwnerBindingUnauthorized) {
		t.Fatalf("foreign binder error = %v", err)
	}
}
