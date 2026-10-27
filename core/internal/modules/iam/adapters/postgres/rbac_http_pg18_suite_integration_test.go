//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMRBACHTTPAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "rbac_http_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMRBACRuntime(t, ctx, database, origin)
			applyIAMRBACHTTP(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			seedIAMRBACHTTPFixtures(t, ctx, database)
			assertIAMRBACHTTPUserLifecycle(t, ctx, database)
			assertIAMRBACHTTPTypedEffective(t, ctx, database)
			assertIAMRBACHTTPACLAndIsolation(t, ctx, database)
			assertIAMRBACHTTPRollback(t, ctx, database, "t_iam_a")
		})
	}
}
