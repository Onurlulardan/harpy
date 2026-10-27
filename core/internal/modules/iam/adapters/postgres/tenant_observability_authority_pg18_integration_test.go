//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestIAMTenantObservabilityAuthorityAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "tenant_observability_"+origin)
			prepareIAMTenantObservabilityAuthority(t, ctx, database, origin)
			applyIAMTenantObservabilityAuthority(t, ctx, database, origin)
			assertIAMTenantObservabilityAuthority(t, ctx, database)
		})
	}
}

func prepareIAMTenantObservabilityAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	applyIAMFoundation(t, ctx, database, origin)
	seedIAMFoundationBase(t, ctx, database)
	applyIAMLifecycleGuards(t, ctx, database, origin)
	applyIAMRBAC(t, ctx, database, origin)
	for _, tenant := range iamFoundationTenants {
		seedIAMRBACCatalog(t, ctx, database, tenant.schema)
		seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
		mustExecIAMTenant(
			t, ctx, database, tenant.schema,
			seedIAMTenantObservabilityRoles, tenant.id,
		)
	}
}

func applyIAMTenantObservabilityAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, 3)
	for _, name := range []string{
		"20260716013433_create_iam_user_has_permission_v1",
		"20260716014004_create_iam_rbac_user_bumps_v1",
		iamTenantObservabilityAuthorityMigration,
	} {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}
