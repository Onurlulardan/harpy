//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"
)

func assertGlobalScopeDenials(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	assertGlobalRejected(t, ctx, coordinator, globalCommandRequestFor(t, 6, 7), ErrGlobalAuthority)
	assertGlobalRejected(t, ctx, coordinator, globalCommandRequestFor(t, 5, 8), ErrGlobalAuthority)
	assertGlobalMembershipMismatch(t, ctx, coordinator, database)
	assertGlobalIdentityMissing(t, ctx, coordinator, database)
}

func assertGlobalRejected(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	request GlobalScopeRequest,
	want error,
) {
	t.Helper()
	called := false
	err := coordinator.InGlobal(
		ctx, request, func(context.Context, GlobalScope) error {
			called = true
			return nil
		},
	)
	if !errors.Is(err, want) || called {
		t.Fatalf("global rejection error=%v want=%v callback=%t", err, want, called)
	}
}

func assertGlobalMembershipMismatch(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	adminExecute(t, ctx, database, `INSERT INTO platform.cutover_group_members
VALUES('global',$1,'billing',2,'forms')`, scopeEnvironmentID)
	defer adminExecute(t, context.Background(), database, `DELETE FROM platform.cutover_group_members
WHERE scope_kind='global' AND scope_id=$1 AND group_id='billing' AND ordinal=2`,
		scopeEnvironmentID)
	assertGlobalRejected(
		t, ctx, coordinator, globalCommandRequestFor(t, 5, 7), ErrGlobalAuthority,
	)
}

func assertGlobalIdentityMissing(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	adminExecute(t, ctx, database, "DELETE FROM platform.environment_identity WHERE singleton")
	defer seedScopeEnvironmentIdentity(t, context.Background(), database)
	assertGlobalRejected(
		t, ctx, coordinator, globalCommandRequestFor(t, 5, 7), ErrGlobalIdentity,
	)
}
