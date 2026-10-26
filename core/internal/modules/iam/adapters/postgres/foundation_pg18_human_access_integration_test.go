//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMPortalBindings = `
INSERT INTO iam_portal_subject_bindings (
    id, user_id, access_mode_id, access_mode, subject_shape, subject_type,
    profile_owner, profile_form_id, profile_form_version_id,
    subject_record_id, identifier_field_id, contract_sha256, state,
    revision, created_at, updated_at, activated_at, suspended_at, retired_at
) VALUES
    ('53000000-0000-4000-8000-000000000001',
     '51000000-0000-4000-8000-000000000003',
     '52000000-0000-4000-8000-000000000004', 'portal',
     'self', 'operator', 'iam', NULL, NULL,
     '51000000-0000-4000-8000-000000000003', NULL,
     pg_catalog.decode(pg_catalog.repeat('51', 32), 'hex'), 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
     '2026-07-16T00:00:00Z', NULL, NULL),
    ('53000000-0000-4000-8000-000000000002',
     '51000000-0000-4000-8000-000000000002',
     '52000000-0000-4000-8000-000000000002', 'portal',
     'form_record', 'employee', 'forms',
     '54000000-0000-4000-8000-000000000001',
     '54000000-0000-4000-8000-000000000002',
     '54000000-0000-4000-8000-000000000003',
     '54000000-0000-4000-8000-000000000004',
     pg_catalog.decode(pg_catalog.repeat('52', 32), 'hex'), 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
     '2026-07-16T00:00:00Z', NULL, NULL),
    ('53000000-0000-4000-8000-000000000003',
     '51000000-0000-4000-8000-000000000001',
     '52000000-0000-4000-8000-000000000005', 'portal',
     'self', 'operator', 'iam', NULL, NULL,
     '51000000-0000-4000-8000-000000000001', NULL,
     pg_catalog.decode(pg_catalog.repeat('53', 32), 'hex'), 'suspended', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:01:00Z',
     NULL, '2026-07-16T00:01:00Z', NULL);`

const insertIAMPortalSetups = `
INSERT INTO iam_portal_access_setups (
    binding_id, user_id, access_mode_id, access_mode, setup_state,
    approval_state, provenance, approval_revision, source_evidence_sha256,
    revision, setup_completed_at, submitted_at, decision_at,
    decision_actor_id, created_at, updated_at
) VALUES
    ('53000000-0000-4000-8000-000000000001',
     '51000000-0000-4000-8000-000000000003',
     '52000000-0000-4000-8000-000000000004', 'portal',
     'not_required', 'not_required', 'v1_adopted', 0,
     pg_catalog.decode(pg_catalog.repeat('61', 32), 'hex'), 1,
     NULL, NULL, NULL, NULL,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z'),
    ('53000000-0000-4000-8000-000000000002',
     '51000000-0000-4000-8000-000000000002',
     '52000000-0000-4000-8000-000000000002', 'portal',
     'completed', 'approved', 'native_v2', 1,
     pg_catalog.decode(pg_catalog.repeat('62', 32), 'hex'), 1,
     '2026-07-16T00:02:00Z', '2026-07-16T00:03:00Z',
     '2026-07-16T00:04:00Z',
     '51000000-0000-4000-8000-000000000001',
     '2026-07-16T00:00:00Z', '2026-07-16T00:04:00Z');`

const insertIAMNeverActivatedPortalSetup = `
INSERT INTO iam_portal_access_setups (
    binding_id, user_id, access_mode_id, access_mode, setup_state,
    approval_state, provenance, approval_revision, source_evidence_sha256,
    revision, created_at, updated_at
) VALUES (
    '53000000-0000-4000-8000-000000000003',
    '51000000-0000-4000-8000-000000000001',
    '52000000-0000-4000-8000-000000000005', 'portal',
    'required', 'not_submitted', 'v1_adopted', 0,
    pg_catalog.decode(pg_catalog.repeat('63', 32), 'hex'), 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:01:00Z'
);`

func assertIAMHumanAccessConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMPortalBindings)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMPortalSetups)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMNeverActivatedPortalSetup)

		err := rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_users SET identity_kind = 'service'
WHERE id = $1`, iamUserOneID)
		assertIAMConstraint(t, err, "23514", "iam_users_identity_kind_check")

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_portal_subject_bindings (
    id, user_id, access_mode_id, access_mode, subject_shape, subject_type,
    profile_owner, subject_record_id, contract_sha256, state, revision,
    created_at, updated_at, retired_at
) VALUES (
    '53000000-0000-4000-8000-000000000010', $1, $2, 'portal',
    'self', 'operator', 'iam', $1,
    pg_catalog.decode(pg_catalog.repeat('71', 32), 'hex'), 'retired', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:01:00Z',
    '2026-07-16T00:00:00Z'
)`, iamUserOneID, iamUserOneConsoleModeID)
		assertIAMConstraint(
			t, err, "23503", "iam_portal_subject_bindings_access_mode_fk",
		)

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_portal_subject_bindings (
    id, user_id, access_mode_id, access_mode, subject_shape, subject_type,
    profile_owner, subject_record_id, contract_sha256, state, revision,
    created_at, updated_at, activated_at
) VALUES (
    '53000000-0000-4000-8000-000000000011', $1, $2, 'portal',
    'form_record', 'employee', 'forms',
    '54000000-0000-4000-8000-000000000003',
    pg_catalog.decode(pg_catalog.repeat('72', 32), 'hex'), 'active', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z'
)`, iamUserThreeID, iamUserThreePortalModeID)
		assertIAMConstraint(
			t, err, "23514", "iam_portal_subject_bindings_subject_shape_fields_check",
		)

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_portal_access_setups
SET setup_state = 'completed', setup_completed_at = NULL
WHERE binding_id = '53000000-0000-4000-8000-000000000001'`)
		assertIAMConstraint(
			t, err, "23514", "iam_portal_access_setups_setup_timestamp_check",
		)

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_portal_access_setups
SET approval_state = 'approved', approval_revision = 1
WHERE binding_id = '53000000-0000-4000-8000-000000000003'`)
		assertIAMConstraint(
			t, err, "23514", "iam_portal_access_setups_state_combination_check",
		)

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_portal_access_setups
SET approval_revision = 0
WHERE binding_id = '53000000-0000-4000-8000-000000000002'`)
		assertIAMConstraint(
			t, err, "23514", "iam_portal_access_setups_revision_check",
		)
	}
}
