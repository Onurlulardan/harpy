//go:build integration

package transaction

import (
	"context"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func assertGlobalScopeCommit(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	binder := globalScopeProbeBinder(t)
	var authority OwnerAuthority
	err := coordinator.InGlobal(
		ctx, globalCommandRequestFor(t, 5, 7),
		func(ctx context.Context, scope GlobalScope) error {
			if scope.ActorID() != uuid.MustParse(globalActorID) ||
				scope.EnvironmentID() != uuid.MustParse(scopeEnvironmentID) ||
				scope.Mode() != Command ||
				scope.OperationID() != "billing.catalog.publish" ||
				!ContextOwnsGlobalScope(ctx, scope) {
				t.Fatal("global command scope identity is not canonical")
			}
			probe, err := BindGlobal(scope, binder)
			if err != nil {
				return err
			}
			searchPath, err := probe.searchPath(ctx)
			if err != nil || !slices.Equal(searchPath, []string{"pg_catalog"}) {
				t.Fatalf("global search_path = %v, %v", searchPath, err)
			}
			authority, err = MintGlobalOwnerAuthority(scope, "billing")
			if err != nil {
				return err
			}
			if !authority.OwnsGlobalScope(scope) {
				t.Fatal("global owner authority does not own exact scope")
			}
			return probe.insert(ctx, "command-commit", "committed")
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if authority.ScopeID() != uuid.Nil || authority.OwnsGlobalScope(
		GlobalScope{state: authority.globalState},
	) {
		t.Fatal("global owner authority escaped commit")
	}
	var value string
	if err := database.admin.QueryRow(
		ctx, "SELECT value FROM platform.billing_global_probe WHERE id='command-commit'",
	).Scan(&value); err != nil || value != "committed" {
		t.Fatalf("global commit value = %q, %v", value, err)
	}
}
