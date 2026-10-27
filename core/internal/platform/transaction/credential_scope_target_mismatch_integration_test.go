//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func assertCredentialTargetSubstitutionRejected(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
) {
	t.Helper()
	lookup, err := NewTenantRouteCredentialLookup(
		"iam.credential.substitution", CredentialPasswordReset,
		uuid.MustParse(tenantA), [32]byte{1},
	)
	if err != nil {
		t.Fatal(err)
	}
	binder, err := NewCredentialResolverBinder(
		func(database postgres.BoundExecutor) CredentialTargetResolver {
			return credentialIntegrationResolver{
				database: database, tenantID: uuid.MustParse(tenantB), generation: 1,
			}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	err = coordinator.InCredentialTenant(
		ctx, CredentialScopeRequest{Lookup: lookup, Resolver: binder},
		func(context.Context, CredentialTenantScope) error {
			called = true
			return nil
		},
	)
	if !errors.Is(err, ErrCredentialTarget) || called {
		t.Fatalf("target substitution: called=%v err=%v", called, err)
	}
}
