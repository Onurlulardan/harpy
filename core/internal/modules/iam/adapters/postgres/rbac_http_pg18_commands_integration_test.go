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

const createIAMRBACHTTPAssignment = `
WITH mutation AS MATERIALIZED (
    SELECT * FROM iam_rbac_assignment_create_v1(
        pg_catalog.current_schema()::name, $1::uuid, $2::bytea,
        $3::uuid, '91000000-0000-4000-8000-000000000001',
        $4::text, $5::uuid, $6::jsonb,
        '51000000-0000-4000-8000-000000000001',
        '93000000-0000-4000-8000-000000000001', $7::bigint, $8::bigint,
		$9::bytea, pg_catalog.transaction_timestamp(), true, false, false
    )
), audit AS MATERIALIZED (
    SELECT iam_rbac_audit_write_v1(
        pg_catalog.current_schema()::name, $1::uuid,
        '51000000-0000-4000-8000-000000000001', mutation.result_id,
        $2::bytea, 'iam.rbac.assignment.create',
        'iam.rbac.assignment.created', mutation.result_revision,
        mutation.authorization_clock
    ) AS written FROM mutation WHERE NOT mutation.replayed
)
SELECT pg_catalog.concat_ws('|', mutation.result_id::text,
    mutation.result_revision::text, mutation.authorization_clock::text,
    mutation.replayed::text, coalesce(audit.written, true)::text)
FROM mutation LEFT JOIN audit ON true`

const replayIAMRBACHTTPAssignment = `
SELECT pg_catalog.concat_ws('|', result_id::text, result_revision::text,
    authorization_clock::text, replayed::text)
FROM iam_rbac_assignment_create_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea,
    $3::uuid, '91000000-0000-4000-8000-000000000001',
    $4::text, $5::uuid, $6::jsonb,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', $7::bigint, $8::bigint,
	$9::bytea, pg_catalog.transaction_timestamp(), true, false, false
)`

func mustCreateIAMRBACHTTPAssignment(
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
	marker byte,
) string {
	t.Helper()
	token, authorization := iamRBACActorVersions(t, ctx, database, schema)
	items := fmt.Sprintf(
		`[{"ordinal":1,"action_id":"%s","permission_id":"%s","binding_id":"%s"}]`,
		actionID, permissionID, bindingID,
	)
	purpose := sha256.Sum256([]byte(
		"harpy/iam/rbac-step-up/v1|iam.rbac.assignment.create|",
	))
	return mustRunIAMAtomicText(
		t, ctx, database, schema, createIAMRBACHTTPAssignment,
		operationID, bytes.Repeat([]byte{marker}, 32), groupID,
		target, targetID, items, token, authorization, purpose[:],
	)
}
