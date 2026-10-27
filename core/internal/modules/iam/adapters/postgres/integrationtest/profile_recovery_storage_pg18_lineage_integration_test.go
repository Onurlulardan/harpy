//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertRecoveryLineageConstraints(
	t *testing.T, ctx context.Context, db publicationDatabase, parentID uuid.UUID,
) {
	t.Helper()
	assertRecoveryParentSchemaStates(t, ctx, db, parentID)
	expectRecoveryOwnerError(t, ctx, db,
		recoveryInvalidSchemaStateSQL+";"+recoveryStorageChildInsertSQL,
		"23514", parentID, recoveryStoragePlanID(t, recoveryStorageBadChild))
	expectRecoveryOwnerError(t, ctx, db, recoveryStorageChildInsertSQL,
		"23514", parentID, recoveryStoragePlanID(t, recoveryStorageBadChild))
	childID := recoveryStoragePlanID(t, recoveryStorageChild)
	createRecoveryStorageChild(t, ctx, db, parentID, childID)
	assertRecoveryZeroPageChild(t, ctx, db, parentID, parentID, childID, 1)
	expectRecoveryOwnerError(t, ctx, db, recoveryStorageChildInsertSQL,
		"23505", parentID, recoveryStoragePlanID(t, recoveryStorageFork))
	expectRecoveryOwnerError(t, ctx, db,
		`UPDATE iam_profile_backfill_plans
SET plan_sha256=pg_catalog.sha256(pg_catalog.convert_to('mutated','UTF8'))
WHERE id=$1`, "55000", parentID)
	expectRecoveryOwnerError(t, ctx, db,
		`UPDATE iam_profile_backfill_plans SET recovery_of_plan_id=id WHERE id=$1`,
		"55000", childID)
	fenceRecoveryStoragePlan(t, ctx, db, childID)
	grandchildID := recoveryStoragePlanID(t, recoveryStorageGrandchild)
	createRecoveryStorageChild(t, ctx, db, childID, grandchildID)
	assertRecoveryZeroPageChild(t, ctx, db, parentID, childID, grandchildID, 2)
	expectRecoveryOwnerError(t, ctx, db, recoveryStorageChildInsertSQL,
		"23505", childID, recoveryStoragePlanID(t, recoveryStorageFork))
}

const recoveryInvalidSchemaStateSQL = `UPDATE iam_profile_schema_versions
SET state='retired',retired_at=pg_catalog.now(),revision=revision+1,
updated_at=pg_catalog.now()
WHERE id=(SELECT source_schema_id FROM iam_profile_backfill_plans WHERE id=$1);
UPDATE iam_profile_schema_versions
SET state='active',activated_at=pg_catalog.now(),revision=revision+1,
updated_at=pg_catalog.now()
WHERE id=(SELECT target_schema_id FROM iam_profile_backfill_plans WHERE id=$1)`

func assertRecoveryParentSchemaStates(
	t *testing.T, ctx context.Context, db publicationDatabase, parentID uuid.UUID,
) {
	t.Helper()
	var sourceState, targetState string
	err := db.admin.QueryRow(ctx, `SELECT source.state,target.state
FROM t_profile_publication.iam_profile_backfill_plans AS plan
JOIN t_profile_publication.iam_profile_schema_versions AS source
ON source.id=plan.source_schema_id
JOIN t_profile_publication.iam_profile_schema_versions AS target
ON target.id=plan.target_schema_id WHERE plan.id=$1`, parentID).
		Scan(&sourceState, &targetState)
	if err != nil || sourceState != "active" || targetState != "draft" {
		t.Fatalf("recovery parent schemas = %s/%s/%v", sourceState, targetState, err)
	}
}

func assertRecoveryZeroPageChild(
	t *testing.T, ctx context.Context, db publicationDatabase,
	rootID, parentID, childID uuid.UUID, sequence int64,
) {
	t.Helper()
	var parent, root uuid.UUID
	var actualSequence int64
	var pageCount, pages, checkpoint, mappings int
	err := db.admin.QueryRow(ctx, `SELECT plan.recovery_of_plan_id,
plan.recovery_root_plan_id,plan.recovery_sequence,plan.page_count,
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_pages
 WHERE plan_id=plan.id),
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_checkpoints
 WHERE plan_id=plan.id AND completed_pages=0 AND converted_count=0
 AND error_count=0 AND last_page_number IS NULL),
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_mappings
 WHERE plan_id=plan.id)
FROM t_profile_publication.iam_profile_backfill_plans AS plan WHERE plan.id=$1`,
		childID).Scan(&parent, &root, &actualSequence, &pageCount,
		&pages, &checkpoint, &mappings)
	if err != nil || parent != parentID || root != rootID || actualSequence != sequence ||
		pageCount != 0 || pages != 0 || checkpoint != 1 || mappings != 2 {
		t.Fatalf("zero-page recovery = %s/%s/%d/%d/%d/%d/%d/%v",
			parent, root, actualSequence, pageCount, pages, checkpoint, mappings, err)
	}
}
