//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMAuthChallengeConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	tests := []struct {
		statement  string
		code       string
		constraint string
	}{
		{`UPDATE iam_auth_challenges
SET session_id = '81000000-0000-4000-8000-000000000002'
WHERE id = '86000000-0000-4000-8000-000000000003'`,
			"23503", "iam_auth_challenges_session_fk"},
		{`UPDATE iam_auth_challenges
SET authenticator_id = '84000000-0000-4000-8000-000000000002'
WHERE id = '86000000-0000-4000-8000-000000000001'`,
			"23503", "iam_auth_challenges_authenticator_fk"},
		{`UPDATE iam_auth_challenges
SET origin_sha256 = pg_catalog.decode(pg_catalog.repeat('a1', 31), 'hex')
WHERE id = '86000000-0000-4000-8000-000000000001'`,
			"23514", "iam_auth_challenges_origin_context_check"},
		{`UPDATE iam_auth_challenges SET origin_policy_version = NULL
WHERE id = '86000000-0000-4000-8000-000000000001'`,
			"23514", "iam_auth_challenges_origin_context_check"},
		{`UPDATE iam_auth_challenges SET rp_id_sha256 = NULL
WHERE id = '86000000-0000-4000-8000-000000000001'`,
			"23514", "iam_auth_challenges_origin_context_check"},
		{`UPDATE iam_auth_challenges
SET expires_at = issued_at + INTERVAL '5 minutes 1 second'
WHERE id = '86000000-0000-4000-8000-000000000003'`,
			"23514", "iam_auth_challenges_time_check"},
		{`UPDATE iam_auth_challenges
SET state = 'cancelled', terminal_at = issued_at, attempt_count = 6
WHERE id = '86000000-0000-4000-8000-000000000003'`,
			"23514", "iam_auth_challenges_attempt_check"},
		{`UPDATE iam_auth_challenges SET attempt_count = 5
WHERE id = '86000000-0000-4000-8000-000000000003'`,
			"23514", "iam_auth_challenges_lifecycle_check"},
		{`UPDATE iam_auth_challenges SET step_up_purpose_sha256 = NULL
WHERE id = '86000000-0000-4000-8000-000000000003'`,
			"23514", "iam_auth_challenges_purpose_shape_check"},
		{`UPDATE iam_auth_challenges
SET context_sha256 = pg_catalog.decode(pg_catalog.repeat('a2', 31), 'hex')
WHERE id = '86000000-0000-4000-8000-000000000003'`,
			"23514", "iam_auth_challenges_secret_check"},
		{`UPDATE iam_auth_challenges
SET challenge_hmac = pg_catalog.decode(pg_catalog.repeat('91', 32), 'hex')
WHERE id = '86000000-0000-4000-8000-000000000002'`,
			"23505", "iam_auth_challenges_secret_key"},
	}
	for _, test := range tests {
		t.Run(schema+"/"+test.constraint, func(t *testing.T) {
			err := rejectIAMTenant(t, ctx, database, schema, test.statement)
			assertIAMConstraint(t, err, test.code, test.constraint)
		})
	}
}
