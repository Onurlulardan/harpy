//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMRefreshLocator = `
INSERT INTO platform.iam_refresh_locators (
    selector, tenant_id, token_id, family_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation, authority_domain,
    cutover_generation, source_revision, source_sha256,
    source_token_state, issued_at, projected_at,
    family_absolute_expires_at, retain_until
) VALUES (
    pg_catalog.decode(pg_catalog.repeat($1, 16), 'hex'), $2, $3, $4,
    'tenant', $2, 'iam_system_resources', 1, 'iam', 1, 1,
    pg_catalog.decode(pg_catalog.repeat('81', 32), 'hex'),
    'active', '2026-07-16T00:00:00Z', '2026-07-16T00:01:00Z',
    '2026-10-14T00:00:00Z',
    '2026-11-01T00:00:00Z'
)`

func assertIAMPlatformAuthConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for index, tenant := range iamFoundationTenants {
		hexByte := "82"
		if index == 1 {
			hexByte = "83"
		}
		if _, err := database.ExecContext(
			ctx, insertIAMRefreshLocator, hexByte, tenant.id,
			"81000000-0000-4000-8000-000000000001",
			"82000000-0000-4000-8000-000000000001",
		); err != nil {
			t.Fatalf("insert %s IAM refresh locator: %v", tenant.schema, err)
		}
	}
	_, err := database.ExecContext(ctx, `
UPDATE platform.iam_refresh_locators
SET selector = pg_catalog.decode('84', 'hex')
WHERE tenant_id = $1`, iamFoundationTenants[0].id)
	assertIAMConstraint(t, err, "23514", "iam_refresh_locators_selector_check")
	_, err = database.ExecContext(ctx, `
UPDATE platform.iam_refresh_locators
SET retain_until = family_absolute_expires_at + INTERVAL '23 hours'
WHERE tenant_id = $1`, iamFoundationTenants[0].id)
	assertIAMConstraint(t, err, "23514", "iam_refresh_locators_time_check")

	if _, err := database.ExecContext(ctx, `
INSERT INTO platform.iam_auth_throttle_buckets (
    bucket_scope_kind, scope_id, purpose, dimension, key_id, bucket_hmac,
    environment_id, tenant_id, authority_scope_kind, authority_scope_id,
    authority_group_id, authority_group_generation, authority_domain,
    cutover_generation, policy_version, window_start, window_end,
    attempt_count, blocked_until, last_attempt_at, revision, retain_until
) VALUES (
    'tenant', $1, 'login', 'identifier_ip', 'auth-hmac-v1',
    pg_catalog.decode(pg_catalog.repeat('91', 32), 'hex'),
    NULL, $1, 'tenant', $1, 'iam_system_resources', 1, 'iam', 1, 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:05:00Z', 2, NULL,
    '2026-07-16T00:01:00Z', 1, '2026-07-17T00:05:00Z'
)`, iamFoundationTenants[0].id); err != nil {
		t.Fatalf("insert tenant IAM throttle: %v", err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO platform.iam_auth_throttle_buckets (
    bucket_scope_kind, scope_id, purpose, dimension, key_id, bucket_hmac,
    environment_id, tenant_id, authority_scope_kind, authority_scope_id,
    authority_group_id, authority_group_generation, authority_domain,
    cutover_generation, policy_version, window_start, window_end,
    attempt_count, blocked_until, last_attempt_at, revision, retain_until
) VALUES (
    'environment', $1, 'login', 'tenant_slug', 'auth-hmac-v1',
    pg_catalog.decode(pg_catalog.repeat('92', 32), 'hex'),
    $1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:05:00Z', 1, NULL,
    '2026-07-16T00:01:00Z', 1, '2026-07-17T00:05:00Z'
)`, "30000000-0000-4000-8000-000000000001"); err != nil {
		t.Fatalf("insert environment IAM throttle: %v", err)
	}
	_, err = database.ExecContext(ctx, `
UPDATE platform.iam_auth_throttle_buckets
SET tenant_id = $1
WHERE bucket_scope_kind = 'environment'`, iamFoundationTenants[0].id)
	assertIAMConstraint(
		t, err, "23514", "iam_auth_throttle_buckets_environment_shape_check",
	)

	var plaintextColumns int
	if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM information_schema.columns
WHERE table_schema = 'platform'
  AND table_name IN ('iam_refresh_locators', 'iam_auth_throttle_buckets')
  AND column_name IN ('email', 'identifier', 'tenant_slug', 'ip_address', 'secret')`).
		Scan(&plaintextColumns); err != nil {
		t.Fatalf("inspect platform IAM plaintext columns: %v", err)
	}
	if plaintextColumns != 0 {
		t.Fatalf("platform IAM auth tables expose %d plaintext columns", plaintextColumns)
	}
}
