//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
)

const retireIAMRBACHTTPAssignment = `
WITH mutation AS MATERIALIZED (
    SELECT * FROM iam_rbac_assignment_retire_v1(
        pg_catalog.current_schema()::name, $1::uuid, $2::bytea,
        $3::uuid, $4::bigint,
        '51000000-0000-4000-8000-000000000001',
        '93000000-0000-4000-8000-000000000001', $5::bigint, $6::bigint,
		$7::bytea, pg_catalog.transaction_timestamp(), true, false, false
    )
), audit AS MATERIALIZED (
    SELECT iam_rbac_audit_write_v1(
        pg_catalog.current_schema()::name, $1::uuid,
        '51000000-0000-4000-8000-000000000001', mutation.result_id,
        $2::bytea, 'iam.rbac.assignment.retire',
        'iam.rbac.assignment.retired', mutation.result_revision,
        mutation.authorization_clock
    ) AS written FROM mutation
)
SELECT pg_catalog.concat_ws('|', mutation.result_id::text,
    mutation.result_revision::text, mutation.authorization_clock::text,
    mutation.replayed::text, audit.written::text)
FROM mutation CROSS JOIN audit`

func mustRetireIAMRBACHTTPAssignment(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	operationID string,
	groupID string,
	expectedRevision int64,
) string {
	t.Helper()
	token, authorization := iamRBACActorVersions(t, ctx, database, schema)
	purpose := sha256.Sum256([]byte(
		"harpy/iam/rbac-step-up/v1|iam.rbac.assignment.retire|" + groupID,
	))
	return mustRunIAMAtomicText(
		t, ctx, database, schema, retireIAMRBACHTTPAssignment,
		operationID, bytes.Repeat([]byte{0xc3}, 32), groupID,
		expectedRevision, token, authorization, purpose[:],
	)
}
