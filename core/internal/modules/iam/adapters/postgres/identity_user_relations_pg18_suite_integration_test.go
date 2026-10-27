//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestIAMUserRelationReplaceAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "user_relations_"+origin)
			prepareIAMUserRelationsPG18(t, ctx, database, origin)
			t.Run("create_update_versions", func(t *testing.T) {
				assertIAMUserRelationCreateUpdate(t, ctx, database)
			})
			t.Run("security", func(t *testing.T) {
				assertIAMUserRelationSecurity(t, ctx, database)
			})
			t.Run("replay_conflict", func(t *testing.T) {
				assertIAMUserRelationReplayConflict(t, ctx, database)
			})
			t.Run("statement_rollback", func(t *testing.T) {
				assertIAMUserRelationStatementRollback(t, ctx, database)
			})
			t.Run("tenant_isolation", func(t *testing.T) {
				assertIAMUserRelationTenantIsolation(t, ctx, database)
			})
			if origin == "v1_adopted" {
				assertIAMV1SnapshotUnchanged(t, ctx, database)
			}
		})
	}
}

func prepareIAMUserRelationsPG18(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	applyIAMFoundation(t, ctx, database, origin)
	prepareIAMIdentityRuntimeAuthority(t, ctx, database, origin)
	applyIAMLifecycleGuards(t, ctx, database, origin)
	applyIAMRBAC(t, ctx, database, origin)
	applyIAMIdentityPermissionResolver(t, ctx, database, origin)
	seedIAMFoundationBase(t, ctx, database)
	applyIAMProfileRevisionMigrations(t, ctx, database, origin)
	applyIAMIdentityRuntime(t, ctx, database, origin)
	seedIAMIdentityRuntime(t, ctx, database)
	seedIAMUserRelationsPG18(t, ctx, database)
}
