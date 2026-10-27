//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMRBACAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "rbac_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			assertIAMRBACCatalog(t, ctx, database)
			assertIAMFoundationHasNoRuntimeDML(t, ctx, database)
			for _, tenant := range iamFoundationTenants {
				seedIAMRBACCatalog(t, ctx, database, tenant.schema)
				seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
				t.Run(tenant.schema, func(t *testing.T) {
					assertIAMRBACRelationalConstraints(t, ctx, database, tenant.schema)
					assertIAMRBACProtectedAdminGuards(t, ctx, database, tenant)
				})
			}
		})
	}
}
