//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func applyIAMPermissionInvitation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0,
		len(iamPermissionInvitationMigrationNames))
	for _, name := range iamPermissionInvitationMigrationNames {
		if name >= "20260716015200" {
			continue
		}
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func applyIAMInvitationHTTPRuntime(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	prerequisites := []string{
		"20260716013701_create_iam_identity_profile_assert_v1",
		"20260716014612_create_iam_hmac_key_id_valid_v1",
	}
	documents := make([]iamFoundationDocument, 0, 16)
	for _, name := range prerequisites {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, name := range iamPermissionInvitationMigrationNames {
		if name < "20260716015200" {
			continue
		}
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func applyIAMInvitationPlatform(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
ALTER TABLE platform.cutover_groups
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2';
ALTER TABLE platform.domain_cutovers
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2';
GRANT USAGE ON SCHEMA platform TO harpy_runtime;`)
	if err != nil {
		t.Fatalf("prepare invitation platform authority: %v", err)
	}
	names := []string{
		"20260716012020_create_iam_tenant_authority_assertion_v1",
		"20260716014602_create_iam_hmac_key_id_valid_v1",
		"20260716015300_create_iam_identity_directory_upsert_v2",
	}
	documents := make([]iamFoundationDocument, 0, len(names))
	for _, name := range names {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMPlatformAuthMigration(t, origin, name),
		})
	}
	applyIAMPlatformDocuments(t, ctx, database, documents)
}
