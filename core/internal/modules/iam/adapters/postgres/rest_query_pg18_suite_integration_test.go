//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMRestQueryAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "rest_query_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			applyIAMRestQueryMigrations(t, ctx, database, origin)
			seedIAMRestQueryTenants(t, ctx, database)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			assertIAMRestQueryACL(t, ctx, database)
			assertIAMRestQueryScopeSecurity(t, ctx, database)
			assertIAMRestQueryTenant(t, ctx, database, "t_iam_a", "tenant-a")
			assertIAMRestQueryTenant(t, ctx, database, "t_iam_b", "tenant-b")
		})
	}
}
