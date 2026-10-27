//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"
)

const updateIAMRBACHTTPAssignment = `
WITH mutation AS MATERIALIZED (
    SELECT * FROM iam_rbac_assignment_update_v1(
        pg_catalog.current_schema()::name, $1::uuid, $2::bytea,
        $3::uuid, '91000000-0000-4000-8000-000000000001',
        $4::text, $5::uuid, $6::jsonb, $7::bigint,
        '51000000-0000-4000-8000-000000000001',
        '93000000-0000-4000-8000-000000000001', $8::bigint, $9::bigint,
		$10::bytea, pg_catalog.transaction_timestamp(), true, false, false
    )
), audit AS MATERIALIZED (
    SELECT iam_rbac_audit_write_v1(
        pg_catalog.current_schema()::name, $1::uuid,
        '51000000-0000-4000-8000-000000000001', mutation.result_id,
        $2::bytea, 'iam.rbac.assignment.update',
        'iam.rbac.assignment.updated', mutation.result_revision,
        mutation.authorization_clock
    ) AS written FROM mutation
)
SELECT pg_catalog.concat_ws('|', mutation.result_id::text,
    mutation.result_revision::text, mutation.authorization_clock::text,
    mutation.replayed::text, audit.written::text)
FROM mutation CROSS JOIN audit`

func mustUpdateIAMRBACHTTPAssignment(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	operationID string,
	groupID string,
	target string,
	targetID string,
	actionID string,
	permissionID string,
	bindingID string,
	expectedRevision int64,
) string {
	t.Helper()
	token, authorization := iamRBACActorVersions(t, ctx, database, schema)
	items := fmt.Sprintf(
		`[{"ordinal":1,"action_id":"%s","permission_id":"%s","binding_id":"%s"}]`,
		actionID, permissionID, bindingID,
	)
	purpose := sha256.Sum256([]byte(
		"harpy/iam/rbac-step-up/v1|iam.rbac.assignment.update|" + groupID,
	))
	return mustRunIAMAtomicText(
		t, ctx, database, schema, updateIAMRBACHTTPAssignment,
		operationID, bytes.Repeat([]byte{0xb2}, 32), groupID,
		target, targetID, items, expectedRevision, token, authorization, purpose[:],
	)
}
