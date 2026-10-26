//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

const (
	recoveryStorageSecondUser = "51000000-0000-4000-8000-000000000002"
	recoveryStorageBadChild   = "61000000-0000-4000-8000-000000000001"
	recoveryStorageChild      = "61000000-0000-4000-8000-000000000002"
	recoveryStorageFork       = "61000000-0000-4000-8000-000000000003"
	recoveryStorageGrandchild = "61000000-0000-4000-8000-000000000004"
)

func seedRecoveryStorageSecondUser(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	document := []byte(`{"43000000-0000-4000-8000-000000000001":"second@example.com"}`)
	digest := sha256.Sum256(document)
	const statement = `INSERT INTO iam_users(
id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
profile_values,profile_sha256,token_version,authorization_version,revision,
created_at,updated_at,tombstoned_at) VALUES(
$1,'human','active','user',$2,$3::jsonb,$4,1,1,1,
'2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL)`
	execPublicationTenant(t, ctx, db, statement, recoveryStorageSecondUser,
		publicationSourceID, string(document), digest[:])
}

func fenceRecoveryStoragePlan(
	t *testing.T, ctx context.Context, db publicationDatabase, planID uuid.UUID,
) {
	t.Helper()
	const statement = `UPDATE iam_profile_backfill_plans
SET state='fenced',fenced_at=pg_catalog.now(),revision=revision+1,
updated_at=pg_catalog.now() WHERE id=$1`
	execPublicationTenant(t, ctx, db, statement, planID)
}

const recoveryStorageChildInsertSQL = `INSERT INTO iam_profile_backfill_plans(
id,resource_kind,source_schema_id,target_schema_id,converter_version,plan_sha256,
source_count,expected_source_sha256,expected_target_sha256,page_size,page_count,
state,activated_at,fenced_at,revision,created_at,updated_at,recovery_of_plan_id,
recovery_root_plan_id,recovery_sequence,parent_plan_sha256,
target_definition_sha256,inherited_target_count,inherited_target_sha256,
remaining_source_count,remaining_source_sha256,
expected_remaining_target_sha256,initial_rolling_sha256)
SELECT $2,parent.resource_kind,parent.source_schema_id,parent.target_schema_id,
parent.converter_version,pg_catalog.sha256(pg_catalog.convert_to($2::text,'UTF8')),
0,pg_catalog.sha256(''::bytea),pg_catalog.sha256(''::bytea),parent.page_size,0,
'planned',NULL,NULL,1,pg_catalog.now(),pg_catalog.now(),parent.id,
COALESCE(parent.recovery_root_plan_id,parent.id),
COALESCE(parent.recovery_sequence,0)+1,parent.plan_sha256,target.definition_sha256,
(SELECT pg_catalog.count(*) FROM iam_profile_recovery_resources_v1 AS resource
 WHERE resource.resource_kind=parent.resource_kind
 AND resource.schema_version_id=parent.target_schema_id),
pg_catalog.sha256(pg_catalog.convert_to('target:'||$2::text,'UTF8')),
0,pg_catalog.sha256(''::bytea),pg_catalog.sha256(''::bytea),
pg_catalog.sha256(''::bytea)
FROM iam_profile_backfill_plans AS parent
JOIN iam_profile_schema_versions AS target ON target.id=parent.target_schema_id
WHERE parent.id=$1`

const recoveryStorageChildMaterializeSQL = `
INSERT INTO iam_profile_backfill_mappings(
plan_id,ordinal,source_field_id,target_field_id,converter_kind,semantic_sha256)
SELECT $2,ordinal,source_field_id,target_field_id,converter_kind,semantic_sha256
FROM iam_profile_backfill_mappings WHERE plan_id=$1;
INSERT INTO iam_profile_backfill_checkpoints(
plan_id,completed_pages,converted_count,error_count,rolling_sha256,
last_page_number,revision,updated_at)
VALUES($2,0,0,0,pg_catalog.sha256(''::bytea),NULL,1,pg_catalog.now())`

func createRecoveryStorageChild(
	t *testing.T, ctx context.Context, db publicationDatabase,
	parentID, childID uuid.UUID,
) {
	t.Helper()
	execPublicationTenant(t, ctx, db,
		recoveryStorageChildInsertSQL+";"+recoveryStorageChildMaterializeSQL,
		parentID, childID)
}
