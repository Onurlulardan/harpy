//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func prepareIAMIdentityRuntimeAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
ALTER ROLE harpy_v2_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS;
GRANT harpy_runtime TO harpy_v2_runtime;
GRANT USAGE ON SCHEMA platform,t_iam_a,t_iam_b TO harpy_runtime;
ALTER TABLE platform.cutover_groups
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2';
ALTER TABLE platform.domain_cutovers
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2';`)
	if err != nil {
		t.Fatalf("prepare IAM identity runtime authority: %v", err)
	}
	applyIAMPlatformDocuments(t, ctx, database, []iamFoundationDocument{{
		path: "20260716012020_create_iam_tenant_authority_assertion_v1",
		content: readIAMPlatformAuthMigration(t, origin,
			"20260716012020_create_iam_tenant_authority_assertion_v1"),
	}})
}

func applyIAMIdentityRuntime(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	platform := make([]iamFoundationDocument, 0, len(iamIdentityPlatformMigrationNames))
	for _, name := range iamIdentityPlatformMigrationNames {
		platform = append(platform, iamFoundationDocument{
			path: name, content: readIAMPlatformAuthMigration(t, origin, name),
		})
	}
	applyIAMPlatformDocuments(t, ctx, database, platform)
	tenant := make([]iamFoundationDocument, 0, len(iamIdentityTenantMigrationNames))
	for _, name := range iamIdentityTenantMigrationNames {
		tenant = append(tenant, iamFoundationDocument{
			path: name, content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, fixture := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, fixture.schema, tenant)
	}
}

func applyIAMIdentityPermissionResolver(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents := []iamFoundationDocument{{
		path: "20260716013433_create_iam_user_has_permission_v1",
		content: readIAMIdentityMigration(t, origin,
			"20260716013433_create_iam_user_has_permission_v1"),
	}}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}
