//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMLifecycleGuardCatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	var always int
	err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM pg_catalog.pg_trigger AS trigger
JOIN pg_catalog.pg_class AS relation ON relation.oid = trigger.tgrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = $1 AND NOT trigger.tgisinternal
  AND trigger.tgname LIKE 'iam_%_guard' AND trigger.tgenabled = 'A'`, schema).Scan(&always)
	if err != nil {
		t.Fatalf("inspect %s IAM lifecycle triggers: %v", schema, err)
	}
	if always != 54 {
		t.Fatalf("%s IAM ALWAYS lifecycle triggers = %d, want 54", schema, always)
	}
	var hardened, publicExecute int
	err = database.QueryRowContext(ctx, `
WITH routines AS (
    SELECT routine.oid, routine.proacl
    FROM pg_catalog.pg_proc AS routine
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace
    WHERE namespace.nspname = $1 AND routine.proname LIKE 'iam_guard_%_v1'
), public_acl AS (
    SELECT DISTINCT routine.oid
    FROM routines AS routine
    CROSS JOIN LATERAL pg_catalog.aclexplode(
        COALESCE(routine.proacl, pg_catalog.acldefault('f', 0))
    ) AS acl
    WHERE acl.grantee = 0 AND acl.privilege_type = 'EXECUTE'
)
SELECT pg_catalog.count(*) FILTER (
           WHERE routine.prosecdef AND
                 routine.proconfig @> ARRAY['search_path=pg_catalog']::text[]
       ),
       (SELECT pg_catalog.count(*) FROM public_acl)
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace
WHERE namespace.nspname = $1 AND routine.proname LIKE 'iam_guard_%_v1'`, schema).
		Scan(&hardened, &publicExecute)
	if err != nil {
		t.Fatalf("inspect %s IAM lifecycle routines: %v", schema, err)
	}
	if hardened != 5 || publicExecute != 0 {
		t.Fatalf("%s IAM hardened/public routines = %d/%d, want 5/0", schema, hardened, publicExecute)
	}
}

func assertIAMLifecycleGuardMonotonicity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	err := rejectIAMTenantReplica(t, ctx, database, schema, `
UPDATE iam_users SET revision = revision + 2
WHERE id = '51000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514", "iam_users_revision_guard")

	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_users
SET token_version = 2, authorization_version = 1,
    updated_at = '2026-07-16T01:00:00Z', revision = 2
WHERE id = '51000000-0000-4000-8000-000000000002'`)
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_users
SET token_version = 1, authorization_version = 0,
    updated_at = '2026-07-16T01:01:00Z', revision = 3
WHERE id = '51000000-0000-4000-8000-000000000002'`)
	assertIAMConstraint(t, err, "23514", "iam_users_security_version_guard")

	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_webauthn_credentials
SET sign_count = 1, updated_at = '2026-07-16T01:00:00Z', revision = 2
WHERE authenticator_id = '84000000-0000-4000-8000-000000000001'`)
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_webauthn_credentials
SET sign_count = 0, updated_at = '2026-07-16T01:01:00Z', revision = 3
WHERE authenticator_id = '84000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514",
		"iam_webauthn_credentials_sign_count_monotonic_guard")

	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_totp_authenticators
SET last_accepted_time_step = 100, updated_at = '2026-07-16T01:00:00Z', revision = 2
WHERE authenticator_id = '84000000-0000-4000-8000-000000000002'`)
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_totp_authenticators
SET last_accepted_time_step = 99, updated_at = '2026-07-16T01:01:00Z', revision = 3
WHERE authenticator_id = '84000000-0000-4000-8000-000000000002'`)
	assertIAMConstraint(t, err, "23514",
		"iam_totp_authenticators_last_accepted_time_step_monotonic_guard")

	assertIAMLifecycleAttemptCounters(t, ctx, database, schema)
	assertIAMLifecycleImmutableBindings(t, ctx, database, schema)
}

func assertIAMLifecycleAttemptCounters(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	for _, test := range []struct {
		table, key, id, constraint string
	}{
		{"iam_auth_challenges", "id", "86000000-0000-4000-8000-000000000002",
			"iam_auth_challenges_attempt_count_monotonic_guard"},
		{"iam_one_time_tokens", "id", "88000000-0000-4000-8000-000000000002",
			"iam_one_time_tokens_attempt_count_monotonic_guard"},
	} {
		mustExecIAMTenant(t, ctx, database, schema,
			"UPDATE "+test.table+" SET attempt_count = 1, revision = 2 WHERE "+test.key+" = $1", test.id)
		err := rejectIAMTenant(t, ctx, database, schema,
			"UPDATE "+test.table+" SET attempt_count = 0, revision = 3 WHERE "+test.key+" = $1", test.id)
		assertIAMConstraint(t, err, "23514", test.constraint)
	}
	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_refresh_families SET rotation_sequence = 1, revision = 2
WHERE id = '82000000-0000-4000-8000-000000000001'`)
	err := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_refresh_families SET rotation_sequence = 0, revision = 3
WHERE id = '82000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514",
		"iam_refresh_families_rotation_sequence_monotonic_guard")
}

func assertIAMLifecycleImmutableBindings(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	err := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_refresh_tokens
SET auth_time = auth_time - INTERVAL '1 second', revision = 2
WHERE id = '83000000-0000-4000-8000-000000000002'`)
	assertIAMConstraint(t, err, "23514", "iam_refresh_tokens_immutable_guard")
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_one_time_tokens
SET secret_hmac = pg_catalog.decode(pg_catalog.repeat('ee', 32), 'hex'), revision = 2
WHERE id = '88000000-0000-4000-8000-000000000004'`)
	assertIAMConstraint(t, err, "23514", "iam_one_time_tokens_immutable_guard")
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_one_time_tokens SET portal_setup_revision = 2, revision = 2
WHERE id = '88000000-0000-4000-8000-000000000003'`)
	assertIAMConstraint(t, err, "23514", "iam_one_time_tokens_immutable_guard")
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_refresh_tokens SET parent_sequence = 99, revision = 2
WHERE id = '83000000-0000-4000-8000-000000000002'`)
	assertIAMConstraint(t, err, "23514", "iam_refresh_tokens_immutable_guard")
}
