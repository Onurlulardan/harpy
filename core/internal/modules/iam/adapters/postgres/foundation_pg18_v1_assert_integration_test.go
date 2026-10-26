//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMV1SnapshotUnchanged(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
		var tenantID, email, password, roleID, organizationID, status string
		var tokenVersion int
		err := transaction.QueryRowContext(ctx, `
SELECT tenant_id::text, email, password, role_ids[1], organization_id,
       status, token_version
FROM users_form_table
WHERE id = '93000000-0000-4000-8000-000000000001'`).Scan(
			&tenantID, &email, &password, &roleID, &organizationID,
			&status, &tokenVersion,
		)
		if err != nil {
			_ = transaction.Rollback()
			_ = connection.Close()
			t.Fatalf("read %s V1 IAM row: %v", tenant.schema, err)
		}
		var legacyObjects int
		if err := transaction.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = pg_catalog.current_schema()
  AND relation.relname IN (
      'organizations_form_table', 'roles_form_table',
      'users_form_table', 'refresh_tokens'
  ) AND relation.relkind = 'r'`).Scan(&legacyObjects); err != nil {
			_ = transaction.Rollback()
			_ = connection.Close()
			t.Fatalf("read %s V1 IAM catalog: %v", tenant.schema, err)
		}
		var hybridType, subjectType, subjectID, profileSlug, loginID string
		var identifierField sql.NullString
		var setupCompleted sql.NullTime
		var setupRequired bool
		var workflowStatus string
		if err := transaction.QueryRowContext(ctx, `
SELECT account_type, portal_subject_type, portal_subject_id::text,
       portal_profile_form_slug, portal_login_identifier,
       portal_identifier_field, portal_setup_required,
       portal_setup_completed_at, workflow_status
FROM users_form_table
WHERE id = '93000000-0000-4000-8000-000000000002'`).Scan(
			&hybridType, &subjectType, &subjectID, &profileSlug, &loginID,
			&identifierField, &setupRequired, &setupCompleted, &workflowStatus,
		); err != nil {
			_ = transaction.Rollback()
			_ = connection.Close()
			t.Fatalf("read %s V1 hybrid IAM row: %v", tenant.schema, err)
		}
		if err := transaction.Commit(); err != nil {
			_ = connection.Close()
			t.Fatalf("commit %s V1 assertion: %v", tenant.schema, err)
		}
		_ = connection.Close()
		if tenantID != tenant.id || email != "Legacy.User@example.com" ||
			password != "$2a$12$legacy-hash-evidence" ||
			roleID != "92000000-0000-4000-8000-000000000001" ||
			organizationID != "91000000-0000-4000-8000-000000000001" ||
			status != "ACTIVE" || tokenVersion != 7 || legacyObjects != 4 {
			t.Fatalf("%s V1 IAM snapshot changed during expand", tenant.schema)
		}
		if hybridType != "operator_portal" || subjectType != "operator" ||
			subjectID != "93000000-0000-4000-8000-000000000002" ||
			profileSlug != "users" || loginID != "OP-0002" ||
			identifierField.Valid || setupRequired || setupCompleted.Valid ||
			workflowStatus != "DRAFT" {
			t.Fatalf("%s V1 hybrid account changed during expand", tenant.schema)
		}
		var rejectedStatus, rejectedWorkflow string
		var rejectedSetup, rejectedPasswordMissing bool
		if err := database.QueryRowContext(ctx, `
SELECT status, portal_setup_required, password IS NULL, workflow_status
FROM `+tenant.schema+`.users_form_table
WHERE id = '93000000-0000-4000-8000-000000000003'`).Scan(
			&rejectedStatus, &rejectedSetup, &rejectedPasswordMissing,
			&rejectedWorkflow,
		); err != nil {
			t.Fatalf("read %s rejected hybrid row: %v", tenant.schema, err)
		}
		if rejectedStatus != "INACTIVE" || !rejectedSetup ||
			!rejectedPasswordMissing || rejectedWorkflow != "REJECTED" {
			t.Fatalf("%s rejected hybrid account changed during expand", tenant.schema)
		}
	}
}
