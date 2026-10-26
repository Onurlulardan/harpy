//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertScopeOwnerAuthority(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
) {
	t.Helper()
	var authority OwnerAuthority
	err := coordinator.InTenant(
		ctx, scopeRequestFor(t, tenantA, 1),
		func(_ context.Context, scope TenantScope) error {
			var err error
			authority, err = MintOwnerAuthority(scope, "iam")
			if err != nil {
				return err
			}
			if authority.ScopeKind() != "tenant" ||
				authority.ScopeID() != uuid.MustParse(tenantA) ||
				authority.GroupID() != "iam_system_resources" || authority.GroupGeneration() != 1 ||
				authority.Domain() != "iam" || authority.Generation() != 1 ||
				authority.Origin() != WriterOriginV2 || !authority.OwnsScope(scope) {
				t.Fatal("owner authority does not match the live PostgreSQL scope")
			}
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if authority.OwnsScope(TenantScope{state: authority.state}) {
		t.Fatal("owner authority remained usable after commit")
	}
}
