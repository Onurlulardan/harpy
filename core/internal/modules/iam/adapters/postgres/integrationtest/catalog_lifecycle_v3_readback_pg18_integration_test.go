//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func readIAMCatalogAuthorizationState(
	t *testing.T, ctx context.Context, db publicationDatabase,
) iamCatalogAuthorizationState {
	t.Helper()
	const statement = `SELECT component.lifecycle_state,
 count(DISTINCT mapped.binding_id),
 count(DISTINCT mapped.binding_id) FILTER (WHERE binding.state='active'),
 (SELECT clock_value FROM iam_authorization_clock WHERE singleton),
 (SELECT authorization_version FROM iam_users WHERE id=$2)
FROM iam_catalog_components AS component
LEFT JOIN iam_catalog_version_operational_grants AS mapped
  ON mapped.component_id=component.component_id
LEFT JOIN iam_role_permission_bindings AS binding
  ON binding.id=mapped.binding_id
WHERE component.component_id=$1
GROUP BY component.lifecycle_state`
	var value iamCatalogAuthorizationState
	err := queryIAMCatalogTenant(
		t, ctx, db, statement, iamCatalogComponentID, publicationActorID,
	).Scan(
		&value.lifecycle, &value.mappedGrants, &value.activeGrants,
		&value.clock, &value.userVersion,
	)
	if err != nil {
		t.Fatalf("read IAM Catalog authorization state: %v", err)
	}
	return value
}

func assertIAMCatalogAuthorizationState(
	t *testing.T, ctx context.Context, db publicationDatabase,
	lifecycle string, activeGrants int, clock, userVersion int64,
) {
	t.Helper()
	actual := readIAMCatalogAuthorizationState(t, ctx, db)
	if actual.lifecycle != lifecycle || actual.mappedGrants != 5 ||
		actual.activeGrants != activeGrants || actual.clock != clock ||
		actual.userVersion != userVersion {
		t.Fatalf("IAM Catalog authorization state differs: %+v", actual)
	}
}

func insertIAMCatalogExternalBinding(
	t *testing.T, ctx context.Context, db publicationDatabase,
	occurredAt time.Time,
) {
	t.Helper()
	const statement = `INSERT INTO iam_user_permission_bindings(
 id,user_id,permission_id,permission_match_kind,state,revision,
 created_at,updated_at,activated_at
)
SELECT $1::uuid,$2::uuid,selected.permission_id,'exact','active',1,$3,$3,$3
FROM (
 SELECT DISTINCT mapped.permission_id
 FROM iam_catalog_version_operational_grants AS mapped
 JOIN iam_catalog_component_heads AS head
   ON head.component_id=mapped.component_id
  AND head.package_id=mapped.package_id
  AND head.package_version=mapped.package_version
  AND head.release_id=mapped.release_id
  AND head.owner_generation=mapped.owner_generation
  AND head.content_sha256=mapped.content_sha256
 WHERE mapped.component_id=$4
 ORDER BY mapped.permission_id
 LIMIT 1
) AS selected
RETURNING id::text`
	var inserted string
	err := queryIAMCatalogTenant(
		t, ctx, db, statement, iamCatalogExternalBinding,
		publicationActorID, occurredAt, iamCatalogComponentID,
	).Scan(&inserted)
	if err != nil || inserted != iamCatalogExternalBinding {
		t.Fatalf("insert IAM Catalog external binding = %s/%v", inserted, err)
	}
}

func revokeIAMCatalogExternalBinding(
	t *testing.T, ctx context.Context, db publicationDatabase,
	occurredAt time.Time,
) {
	t.Helper()
	const statement = `UPDATE iam_user_permission_bindings
SET state='revoked',revision=revision+1,
    updated_at=$2,revoked_at=$2
WHERE id=$1::uuid AND state='active'
RETURNING state`
	var state string
	err := queryIAMCatalogTenant(
		t, ctx, db, statement, iamCatalogExternalBinding, occurredAt,
	).Scan(&state)
	if err != nil || state != "revoked" {
		t.Fatalf("revoke IAM Catalog external binding = %s/%v", state, err)
	}
}

func assertIAMCatalogBlockerReceiptCount(
	t *testing.T, ctx context.Context, db publicationDatabase, want int,
) {
	t.Helper()
	const statement = `SELECT count(*)
FROM iam_catalog_blocker_receipts
WHERE operation_id=$1 AND component_id=$2`
	var count int
	err := queryIAMCatalogTenant(
		t, ctx, db, statement,
		iamCatalogPurgeOperation, iamCatalogComponentID,
	).Scan(&count)
	if err != nil || count != want {
		t.Fatalf("IAM Catalog blocker receipt count = %d/%v", count, err)
	}
}

func assertIAMCatalogPurged(
	t *testing.T, ctx context.Context, db publicationDatabase,
	wantClock, wantUserVersion int64,
) {
	t.Helper()
	const statement = `SELECT component.lifecycle_state,
 (SELECT count(*) FROM iam_catalog_component_heads WHERE component_id=$1),
 (SELECT count(*) FROM iam_catalog_version_evidence WHERE component_id=$1),
 (SELECT count(*) FROM iam_catalog_version_provisions WHERE component_id=$1),
 (SELECT count(*) FROM iam_catalog_version_operational_grants WHERE component_id=$1),
 (SELECT count(*) FROM iam_catalog_version_operational_resources WHERE component_id=$1),
 (SELECT count(*) FROM iam_catalog_materializations WHERE component_id=$1),
 (SELECT count(*) FROM iam_resources WHERE resource_key IN (
   'module.project','form.project-task','workflow.project-task-review'
 ) AND state='retired'),
 (SELECT count(*) FROM iam_permissions AS permission
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  WHERE resource.resource_key IN (
   'module.project','form.project-task','workflow.project-task-review'
  ) AND permission.state='retired'),
 (SELECT count(*) FROM iam_role_permission_bindings AS binding
  JOIN iam_permissions AS permission ON permission.id=binding.permission_id
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  WHERE resource.resource_key IN (
   'module.project','form.project-task','workflow.project-task-review'
  ) AND binding.state='revoked'),
 (SELECT state FROM iam_user_permission_bindings WHERE id=$2),
 (SELECT clock_value FROM iam_authorization_clock WHERE singleton),
 (SELECT authorization_version FROM iam_users WHERE id=$3),
 (SELECT count(*) FROM iam_catalog_purge_contexts)
FROM iam_catalog_components AS component WHERE component.component_id=$1`
	var state, externalState string
	var head, evidence, provisions, grants, resources, legacy int
	var retiredResources, retiredPermissions, revokedBindings, contexts int
	var clock, userVersion int64
	err := queryIAMCatalogTenant(
		t, ctx, db, statement, iamCatalogComponentID,
		iamCatalogExternalBinding, publicationActorID,
	).Scan(
		&state, &head, &evidence, &provisions, &grants, &resources, &legacy,
		&retiredResources, &retiredPermissions, &revokedBindings,
		&externalState, &clock, &userVersion, &contexts,
	)
	if err != nil || state != "PURGED" || head+evidence+provisions+
		grants+resources+legacy+contexts != 0 || retiredResources != 3 ||
		retiredPermissions != 5 || revokedBindings != 5 ||
		externalState != "revoked" || clock != wantClock ||
		userVersion != wantUserVersion {
		t.Fatalf(
			"IAM Catalog purge readback = %s/%d/%d/%d/%d/%d/%d/%d/%d/%d/%s/%d/%d/%d/%v",
			state, head, evidence, provisions, grants, resources, legacy,
			retiredResources, retiredPermissions, revokedBindings,
			externalState, clock, userVersion, contexts, err,
		)
	}
	assertIAMCatalogBlockerReceiptCount(t, ctx, db, 1)
}
