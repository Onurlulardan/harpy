//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const iamPortalSetupReadBindingID = "76000000-0000-4000-8000-000000000001"

func TestIAMPortalSetupReadAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(
				t, ctx, "portal_setup_read_"+origin,
			)
			prepareIAMPortalSetupReadPG18(t, ctx, database, origin)
			assertIAMPortalSetupReadPG18(t, ctx, database)
		})
	}
}

func prepareIAMPortalSetupReadPG18(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	applyIAMFoundation(t, ctx, database, origin)
	prepareIAMIdentityRuntimeAuthority(t, ctx, database, origin)
	applyIAMLifecycleGuards(t, ctx, database, origin)
	applyIAMRBAC(t, ctx, database, origin)
	applyIAMIdentityPermissionResolver(t, ctx, database, origin)
	applyIAMLifecycleIdentityAuthority(t, ctx, database, origin)
	applyIAMLifecycleRuntime(t, ctx, database, origin)
	seedIAMFoundationBase(t, ctx, database)
	seedIAMIdentityRuntime(t, ctx, database)

	documents := make([]iamFoundationDocument, 0, len(iamPortalSetupReadMigrationNames))
	for _, name := range iamPortalSetupReadMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
	mustExecIAMTenant(t, ctx, database, iamFoundationTenants[0].schema, `
INSERT INTO iam_portal_subject_bindings (
    id,user_id,access_mode_id,access_mode,subject_shape,subject_type,
    profile_owner,profile_form_id,profile_form_version_id,subject_record_id,
    identifier_field_id,contract_sha256,state,revision,created_at,updated_at,
    activated_at,suspended_at,retired_at
) VALUES (
    '76000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000002',
    '52000000-0000-4000-8000-000000000002',
    'portal','form_record','employee','forms',
    '76000000-0000-4000-8000-000000000002',
    '76000000-0000-4000-8000-000000000003',
    '76000000-0000-4000-8000-000000000004',
    '76000000-0000-4000-8000-000000000005',
    pg_catalog.decode(pg_catalog.repeat('61',32),'hex'),
    'active',1,'2026-07-16T00:30:00Z','2026-07-16T00:30:00Z',
    '2026-07-16T00:30:00Z',NULL,NULL
);
INSERT INTO iam_portal_access_setups (
    binding_id,user_id,access_mode_id,access_mode,setup_state,approval_state,
    provenance,approval_revision,source_evidence_sha256,revision,
    setup_completed_at,submitted_at,decision_at,decision_actor_id,
    created_at,updated_at
) VALUES (
    '76000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000002',
    '52000000-0000-4000-8000-000000000002',
    'portal','completed','pending','native_v2',1,
    pg_catalog.decode(pg_catalog.repeat('62',32),'hex'),2,
    '2026-07-16T00:31:00Z','2026-07-16T00:32:00Z',NULL,NULL,
    '2026-07-16T00:30:00Z','2026-07-16T00:32:00Z'
);`)
}

func assertIAMPortalSetupReadPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	tenant := iamFoundationTenants[0]
	actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamUserOneID)
	connection, transaction, err := beginIAMAtomicRuntimeTx(
		ctx, database, tenant.schema,
	)
	if err != nil {
		t.Fatal(err)
	}
	var bindingID, userID, setupState, approvalState string
	var revision int64
	var setupCompletedAt, submittedAt, updatedAt time.Time
	var decisionAt *time.Time
	err = transaction.QueryRowContext(ctx, `
SELECT binding_id,user_id,setup_state,approval_state,result_revision,
       setup_completed_at,submitted_at,decision_at,updated_at
FROM iam_portal_setup_read_v1(
    $1::name,$2::uuid,$3::uuid,$4::uuid,$5::uuid,$6::bigint,$7::bigint
)`, tenant.schema, tenant.id, iamPortalSetupReadBindingID, iamUserOneID,
		iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion,
	).Scan(&bindingID, &userID, &setupState, &approvalState, &revision,
		&setupCompletedAt, &submittedAt, &decisionAt, &updatedAt)
	_ = transaction.Rollback()
	_ = connection.Close()
	if err != nil || bindingID != iamPortalSetupReadBindingID ||
		userID != iamUserTwoID || setupState != "completed" ||
		approvalState != "pending" || revision != 2 || decisionAt != nil ||
		setupCompletedAt.IsZero() || submittedAt.IsZero() || updatedAt.IsZero() {
		t.Fatalf("portal setup read=%s/%s/%s/%s/%d/%v/%v/%v/%v: %v",
			bindingID, userID, setupState, approvalState, revision,
			setupCompletedAt, submittedAt, decisionAt, updatedAt, err)
	}

	_, err = queryIAMPortalSetupRevision(
		ctx, database, tenant, "76000000-0000-4000-8000-000000000099",
		actor,
	)
	assertIAMSQLState(t, err, "P0002")
	stale := actor
	stale.authorizationVersion++
	_, err = queryIAMPortalSetupRevision(
		ctx, database, tenant, iamPortalSetupReadBindingID, stale,
	)
	assertIAMSQLState(t, err, "42501")
	wrongTenant := tenant
	wrongTenant.id = iamFoundationTenants[1].id
	_, err = queryIAMPortalSetupRevision(
		ctx, database, wrongTenant, iamPortalSetupReadBindingID, actor,
	)
	assertIAMSQLState(t, err, "42501")
}

func queryIAMPortalSetupRevision(
	ctx context.Context, database *sql.DB, tenant iamFoundationTenant,
	bindingID string, actor iamIdentityActor,
) (int64, error) {
	connection, transaction, err := beginIAMAtomicRuntimeTx(
		ctx, database, tenant.schema,
	)
	if err != nil {
		return 0, err
	}
	defer connection.Close()
	var revision int64
	err = transaction.QueryRowContext(ctx, `
SELECT result_revision
FROM iam_portal_setup_read_v1(
    $1::name,$2::uuid,$3::uuid,$4::uuid,$5::uuid,$6::bigint,$7::bigint
)`, tenant.schema, tenant.id, bindingID, iamUserOneID, iamIdentityPermissionID,
		actor.tokenVersion, actor.authorizationVersion).Scan(&revision)
	_ = transaction.Rollback()
	return revision, err
}
