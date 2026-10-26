package postgres

import (
	"strings"
	"testing"
)

func TestIAMRefreshLocatorMigrationContract(t *testing.T) {
	source := string(readIAMPlatformAuthMigration(
		t, "fresh", "20260716012000_create_iam_refresh_locators",
	))
	markers := []string{
		"CREATE TABLE platform.iam_refresh_locators (",
		"selector bytea PRIMARY KEY",
		"tenant_id uuid NOT NULL",
		"token_id uuid NOT NULL",
		"family_id uuid NOT NULL",
		"UNIQUE (tenant_id, token_id)",
		"pg_catalog.octet_length(selector) = 16",
		"scope_kind = 'tenant' AND scope_id = tenant_id",
		"authority_group_id = 'iam_system_resources'",
		"authority_domain = 'iam'",
		"authority_group_generation > 0 AND cutover_generation > 0",
		"source_revision > 0",
		"pg_catalog.octet_length(source_sha256) = 32",
		"source_token_state IN ('active', 'consumed', 'revoked', 'expired')",
		"issued_at <= projected_at AND issued_at < family_absolute_expires_at",
		"retain_until >= family_absolute_expires_at + INTERVAL '24 hours'",
		"REFERENCES platform.tenants (id)",
		"REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)",
		"REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)",
		"REFERENCES platform.cutover_group_members (",
		"ON UPDATE NO ACTION ON DELETE RESTRICT",
		"iam_refresh_locators_family_idx",
		"iam_refresh_locators_retention_idx",
	}
	assertPlatformAuthMarkers(t, source, markers)
	if strings.Count(source, "FOREIGN KEY") != 4 {
		t.Fatal("refresh locator must have tenant and exact cutover authority FKs")
	}
	for _, forbidden := range []string{
		"secret", "email_address", "identifier_value", "ip_address",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("refresh locator exposes forbidden field %q", forbidden)
		}
	}
}

func TestIAMAuthThrottleMigrationContract(t *testing.T) {
	source := string(readIAMPlatformAuthMigration(
		t, "fresh", "20260716012010_create_iam_auth_throttle_buckets",
	))
	markers := []string{
		"CREATE TABLE platform.iam_auth_throttle_buckets (",
		"bucket_scope_kind, scope_id, purpose, dimension, key_id, bucket_hmac",
		"bucket_scope_kind IN ('tenant', 'environment')",
		"'login', 'refresh', 'password_reset', 'email_verification'",
		"'portal_setup', 'invitation', 'reauth'",
		"'tenant_slug', 'identifier', 'ip', 'identifier_ip', 'selector'",
		"pg_catalog.octet_length(bucket_hmac) = 32",
		"REFERENCES platform.environment_identity (environment_id)",
		"REFERENCES platform.tenants (id)",
		"REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)",
		"REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)",
		"REFERENCES platform.cutover_group_members (",
		"environment_id IS NULL AND tenant_id IS NOT NULL",
		"authority_scope_kind = 'tenant'",
		"authority_scope_id = scope_id",
		"authority_group_id = 'iam_system_resources'",
		"authority_domain = 'iam'",
		"environment_id IS NOT NULL AND environment_id = scope_id",
		"tenant_id IS NULL",
		"authority_scope_kind IS NULL AND authority_scope_id IS NULL",
		"authority_domain IS NULL AND cutover_generation IS NULL",
		"policy_version > 0 AND attempt_count >= 0 AND revision > 0",
		"window_start < window_end",
		"window_end <= retain_until",
	}
	assertPlatformAuthMarkers(t, source, markers)
	if strings.Count(source, "FOREIGN KEY") != 5 {
		t.Fatal("auth throttle must have environment, tenant and cutover FKs")
	}
	for _, forbidden := range []string{
		"tenant_slug text", "identifier_value", "email_address", "ip_address",
		"raw_value", "plaintext",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("auth throttle exposes forbidden field %q", forbidden)
		}
	}
}

func assertPlatformAuthMarkers(t *testing.T, source string, markers []string) {
	t.Helper()
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("platform auth migration lacks marker %q", marker)
		}
	}
}
