//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMRBACRuntimeAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "rbac_runtime_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMRBACRuntime(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			for _, tenant := range iamFoundationTenants {
				seedIAMRBACCatalog(t, ctx, database, tenant.schema)
				seedIAMRBACRuntimeAuthority(t, ctx, database, tenant.schema)
			}
			assertIAMRBACRuntimeReplayAndIsolation(t, ctx, database)
			assertIAMRBACRuntimeAuditRollback(t, ctx, database, "t_iam_a")
			assertIAMRBACBindingABA(t, ctx, database, "t_iam_a")
			assertIAMRBACImplicationLifecycle(t, ctx, database, "t_iam_b")
			assertIAMRBACMutationRace(t, ctx, database, "t_iam_b")
			seedIAMRBACProtectedPolicy(t, ctx, database, iamFoundationTenants[0])
			assertIAMRBACProtectedPolicyRuntime(t, ctx, database, "t_iam_a")
		})
	}
}
