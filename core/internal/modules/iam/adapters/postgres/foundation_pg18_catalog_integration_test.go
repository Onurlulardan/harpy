//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"slices"
	"testing"
)

func assertIAMFoundationCatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	want := []string{
		"iam_auth_challenges",
		"iam_authenticators",
		"iam_invitations",
		"iam_one_time_tokens",
		"iam_organization_memberships",
		"iam_organizations",
		"iam_password_credentials",
		"iam_portal_access_setups",
		"iam_portal_subject_bindings",
		"iam_profile_field_revisions",
		"iam_profile_fields",
		"iam_profile_schema_versions",
		"iam_recovery_codes",
		"iam_refresh_families",
		"iam_refresh_tokens",
		"iam_roles",
		"iam_sessions",
		"iam_totp_authenticators",
		"iam_user_access_modes",
		"iam_user_identifiers",
		"iam_user_role_assignments",
		"iam_users",
		"iam_webauthn_credentials",
	}
	for _, tenant := range iamFoundationTenants {
		rows, err := database.QueryContext(ctx, `
SELECT tablename
FROM pg_catalog.pg_tables
WHERE schemaname = $1 AND tablename LIKE 'iam\_%' ESCAPE '\'
ORDER BY tablename`, tenant.schema)
		if err != nil {
			t.Fatalf("read %s IAM catalog: %v", tenant.schema, err)
		}
		var got []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				_ = rows.Close()
				t.Fatalf("scan %s IAM catalog: %v", tenant.schema, err)
			}
			got = append(got, name)
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("close %s IAM catalog: %v", tenant.schema, err)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("%s IAM tables = %v, want %v", tenant.schema, got, want)
		}
	}
	platformWant := []string{
		"iam_auth_throttle_buckets",
		"iam_login_directory",
		"iam_refresh_locators",
	}
	rows, err := database.QueryContext(ctx, `
SELECT tablename
FROM pg_catalog.pg_tables
WHERE schemaname = 'platform' AND tablename LIKE 'iam\_%' ESCAPE '\'
ORDER BY tablename`)
	if err != nil {
		t.Fatalf("read platform IAM catalog: %v", err)
	}
	var platformGot []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			t.Fatalf("scan platform IAM catalog: %v", err)
		}
		platformGot = append(platformGot, name)
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("close platform IAM catalog: %v", err)
	}
	if !slices.Equal(platformGot, platformWant) {
		t.Fatalf("platform IAM tables = %v, want %v", platformGot, platformWant)
	}
}
