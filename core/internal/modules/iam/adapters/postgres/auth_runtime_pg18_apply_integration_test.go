//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestIAMAuthRuntimeAppliesAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "auth_runtime_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMAtomicRoutines(t, ctx, database, origin)
			applyIAMAuthRuntime(t, ctx, database, origin)
			assertIAMAuthRuntimeACL(t, ctx, database)
			assertIAMAuthSecurityRuntimeACL(t, ctx, database)
		})
	}
}

func applyIAMAuthRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamAuthRuntimeMigrationNames))
	for _, name := range iamAuthRuntimeMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}
