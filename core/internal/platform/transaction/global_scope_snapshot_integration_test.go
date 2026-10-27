//go:build integration

package transaction

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func assertGlobalPublicSnapshot(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
) {
	t.Helper()
	binder := globalScopeProbeBinder(t)
	err := coordinator.InGlobal(
		ctx, globalPublicRequestFor(t, 5, 7),
		func(ctx context.Context, scope GlobalScope) error {
			if scope.ActorID() != uuid.Nil || scope.Mode() != SnapshotQuery ||
				scope.EnvironmentID() != uuid.MustParse(scopeEnvironmentID) {
				t.Fatal("public global snapshot gained actor or wrong environment")
			}
			if _, err := MintGlobalOwnerAuthority(scope, "billing"); !errors.Is(
				err, ErrScopeOwner,
			) {
				t.Fatalf("public snapshot minted command authority: %v", err)
			}
			probe, err := BindGlobal(scope, binder)
			if err != nil {
				return err
			}
			path, err := probe.searchPath(ctx)
			if err != nil || !slices.Equal(path, []string{"pg_catalog"}) {
				t.Fatalf("public search_path = %v, %v", path, err)
			}
			value, err := probe.value(ctx, "public-seed")
			if err != nil || value != "visible" {
				t.Fatalf("public snapshot value = %q, %v", value, err)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}
