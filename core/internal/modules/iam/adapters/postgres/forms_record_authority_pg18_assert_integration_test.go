//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMFormsRecordAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	schema := iamFoundationTenants[0].schema
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	var resourceID string
	var actionCount, permissionCount int64
	var semantics bool
	err := transaction.QueryRowContext(ctx, `
SELECT resource.id::text,
       pg_catalog.count(*) FILTER (WHERE action.action_key IN
           ('create', 'read', 'update', 'delete', 'import', 'export')),
       pg_catalog.count(*),
       pg_catalog.bool_and(permission.match_kind = 'exact' AND
           NOT permission.is_protected AND permission.state = 'active' AND
           action.state = 'active' AND NOT action.is_protected)
FROM iam_resources AS resource
JOIN iam_permissions AS permission ON permission.resource_id = resource.id
JOIN iam_actions AS action ON action.id = permission.action_id
WHERE resource.resource_key = 'forms.records' AND resource.owner_domain = 'forms'
  AND resource.resource_kind = 'system_resource'
  AND resource.is_protected AND resource.state = 'active'
GROUP BY resource.id`).Scan(
		&resourceID, &actionCount, &permissionCount, &semantics,
	)
	if err != nil || resourceID != "e6000000-0000-4000-8000-000000000001" ||
		actionCount != 6 || permissionCount != 6 || !semantics {
		t.Fatalf("Forms record catalog = %s/%d/%d/%t: %v",
			resourceID, actionCount, permissionCount, semantics, err)
	}
	assertIAMFormsRecordReusedIDs(t, ctx, transaction)
	assertIAMFormsRecordEdges(t, ctx, transaction)
	assertIAMFormsRecordNoAutomaticGrants(t, ctx, transaction)
	var protectedCount int64
	err = transaction.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FROM iam_permissions AS permission
JOIN iam_resources AS resource ON resource.id = permission.resource_id
WHERE resource.resource_key = 'forms.records'
  AND iam_user_has_permission_v1($1, permission.id)`, iamUserOneID,
	).Scan(&protectedCount)
	if err != nil || protectedCount != 6 {
		t.Fatalf("protected tenant-all Forms coverage = %d: %v", protectedCount, err)
	}
}

func assertIAMFormsRecordReusedIDs(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
) {
	t.Helper()
	var readAction, importAction, readPermission, importPermission string
	err := transaction.QueryRowContext(ctx, `
SELECT
  (SELECT id::text FROM iam_actions WHERE action_key = 'read'),
  (SELECT id::text FROM iam_actions WHERE action_key = 'import'),
  (SELECT permission.id::text FROM iam_permissions AS permission
   JOIN iam_actions AS action ON action.id = permission.action_id
   JOIN iam_resources AS resource ON resource.id = permission.resource_id
   WHERE resource.resource_key = 'forms.records' AND action.action_key = 'read'),
  (SELECT permission.id::text FROM iam_permissions AS permission
   JOIN iam_actions AS action ON action.id = permission.action_id
   JOIN iam_resources AS resource ON resource.id = permission.resource_id
   WHERE resource.resource_key = 'forms.records' AND action.action_key = 'import')`,
	).Scan(&readAction, &importAction, &readPermission, &importPermission)
	if err != nil || readAction != iamRBACReadActionID ||
		importAction != "e6000000-0000-4000-8000-000000000105" ||
		readPermission != "e6000000-0000-4000-8000-000000000202" ||
		importPermission != "e6000000-0000-4000-8000-000000000205" {
		t.Fatalf("reused Forms authority IDs = %s/%s/%s/%s: %v",
			readAction, importAction, readPermission, importPermission, err)
	}
}

func assertIAMFormsRecordEdges(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
) {
	t.Helper()
	var total, importCreate, exportRead int64
	var semantics bool
	err := transaction.QueryRowContext(ctx, `
SELECT pg_catalog.count(*),
	       pg_catalog.count(*) FILTER (WHERE source_action.action_key = 'import'
	           AND implied_action.action_key = 'create'
	           AND edge.id = 'e6000000-0000-4000-8000-000000000301'),
	       pg_catalog.count(*) FILTER (WHERE source_action.action_key = 'export'
	           AND implied_action.action_key = 'read'
	           AND edge.id = 'f6000000-0000-4000-8000-000000000302'),
       pg_catalog.bool_and(edge.source_match_kind = 'exact' AND
           edge.implied_match_kind = 'exact' AND edge.state = 'active' AND
           edge.retirement_contract_key = 'forms.records.operation.v1')
FROM iam_permission_implications AS edge
JOIN iam_permissions AS source ON source.id = edge.source_permission_id
JOIN iam_actions AS source_action ON source_action.id = source.action_id
JOIN iam_permissions AS implied ON implied.id = edge.implied_permission_id
JOIN iam_actions AS implied_action ON implied_action.id = implied.action_id
WHERE edge.implication_kind = 'forms_record_operation_v1'`,
	).Scan(&total, &importCreate, &exportRead, &semantics)
	if err != nil || total != 2 || importCreate != 1 || exportRead != 1 || !semantics {
		t.Fatalf("Forms record edges = %d/%d/%d/%t: %v",
			total, importCreate, exportRead, semantics, err)
	}
}

func assertIAMFormsRecordNoAutomaticGrants(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
) {
	t.Helper()
	var users, roles, organizations, assignments int64
	err := transaction.QueryRowContext(ctx, `
SELECT (SELECT pg_catalog.count(*) FROM iam_user_permission_bindings),
       (SELECT pg_catalog.count(*) FROM iam_role_permission_bindings),
       (SELECT pg_catalog.count(*) FROM iam_organization_permission_bindings),
       (SELECT pg_catalog.count(*) FROM iam_user_role_assignments)`,
	).Scan(&users, &roles, &organizations, &assignments)
	if err != nil || users != 0 || roles != 1 || organizations != 0 || assignments != 1 {
		t.Fatalf("unexpected automatic grants = %d/%d/%d/%d: %v",
			users, roles, organizations, assignments, err)
	}
}

func assertIAMFormsRecordImplicationAuthorization(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	schema := iamFoundationTenants[0].schema
	mustExecIAMTenant(t, ctx, database, schema, insertIAMRBACUserBinding,
		"e6000000-0000-4000-8000-000000000701", iamUserTwoID,
		"e6000000-0000-4000-8000-000000000205", "exact")
	mustExecIAMTenant(t, ctx, database, schema, insertIAMRBACUserBinding,
		"e6000000-0000-4000-8000-000000000702", iamUserTwoID,
		"f6000000-0000-4000-8000-000000000206", "exact")
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	var createAllowed, readAllowed bool
	err := transaction.QueryRowContext(ctx, `
SELECT iam_user_has_permission_v1($1,
	           'e6000000-0000-4000-8000-000000000201'),
       iam_user_has_permission_v1($1,
           'e6000000-0000-4000-8000-000000000202')`, iamUserTwoID,
	).Scan(&createAllowed, &readAllowed)
	if err != nil || !createAllowed || !readAllowed {
		t.Fatalf("Forms record implications authorize = %t/%t: %v",
			createAllowed, readAllowed, err)
	}
}
