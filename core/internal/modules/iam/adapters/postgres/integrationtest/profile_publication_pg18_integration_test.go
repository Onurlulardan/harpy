//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/google/uuid"
)

func TestProfilePublicationApplicationAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			db := startPublicationPG18(t, ctx, origin)
			preparePublicationCatalog(t, ctx, db)
			seedPublicationV1Snapshot(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "tenant", "foundation"), true)
			source := publicationSourceSchema(t)
			seedPublicationSource(t, ctx, db, source)
			seedPublicationCurrentSource(t, ctx, db)
			seedPublicationEmptyMappingSource(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "tenant", "runtime"), true)
			seedPublicationAuthority(t, ctx, db)
			seedPublicationEmptyMappingAuthority(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "platform", "platform"), false)
			applyPublicationDocuments(t, ctx, db,
				publicationAsyncDocuments(t, origin), false)
			grantPublicationRuntimeAccess(t, ctx, db)
			assertProfileRecoveryACL(t, ctx, db)
			seedPublicationStartup(t, ctx, db, publicationMigrationDigest(t))
			runtime := openPublicationRuntime(t, ctx, db)
			assertPublicationRuntimeAuthority(t, ctx, db, runtime)
			assertProfilePublicationLifecycle(t, ctx, db, runtime)
		})
	}
}

func assertProfilePublicationLifecycle(
	t *testing.T, ctx context.Context, db publicationDatabase, runtime publicationRuntime,
) {
	t.Helper()
	claims := publicationClaims(t, db.tenant)
	assertOptionalPublicationLeavesHistoricalRow(t, ctx, db)
	invalid, err := runtime.application.PublishSchema(
		ctx, claims, publicationInvalidBackfillInput(),
	)
	if !errors.Is(err, iammodule.ErrProfileHTTPConflict) ||
		invalid != (iammodule.ProfileSchemaMutationView{}) {
		t.Fatalf("invalid-row publication result = %#v/%v", invalid, err)
	}
	assertPublicationRolledBack(t, ctx, db)
	setPublicationJobInsert(t, ctx, db, false)
	failed, err := runtime.application.PublishSchema(
		ctx, claims, publicationBackfillInput(),
	)
	if !errors.Is(err, iammodule.ErrProfileHTTPUnavailable) ||
		failed != (iammodule.ProfileSchemaMutationView{}) {
		t.Fatalf("enqueue failure result = %#v/%v", failed, err)
	}
	assertPublicationRolledBack(t, ctx, db)
	setPublicationJobInsert(t, ctx, db, true)
	planned, err := runtime.application.PublishSchema(
		ctx, claims, publicationBackfillInput(),
	)
	if err != nil || planned.ID != uuid.MustParse(publicationTargetV3ID) ||
		planned.BackfillPlanID == nil || planned.PublicationState != "planned" ||
		planned.Replayed {
		t.Fatalf("planned publication = %#v/%v", planned, err)
	}
	replayed, err := runtime.application.PublishSchema(
		ctx, claims, publicationBackfillInput(),
	)
	if err != nil || replayed.BackfillPlanID == nil ||
		*replayed.BackfillPlanID != *planned.BackfillPlanID ||
		!replayed.Replayed || replayed.PublicationState != "planned" {
		t.Fatalf("publication replay = %#v/%v", replayed, err)
	}
	assertPublicationPlanAndFirstJob(t, ctx, db, *planned.BackfillPlanID)
	assertPublicationHistoricalNormalWrite(
		t, ctx, db, runtime, *planned.BackfillPlanID,
	)
	executePublicationProfileWorker(t, ctx, db, *planned.BackfillPlanID)
	assertPublicationWorkerActivation(t, ctx, db, runtime, *planned.BackfillPlanID)
	assertPublicationEmptyMappingLifecycle(t, ctx, db, runtime)
}

func assertOptionalPublicationLeavesHistoricalRow(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	var active, row uuid.UUID
	err := db.admin.QueryRow(ctx, `SELECT active.id,usr.profile_schema_version_id
FROM t_profile_publication.iam_profile_schema_versions AS active
JOIN t_profile_publication.iam_users AS usr ON usr.id=$1
WHERE active.resource_kind='user' AND active.state='active'`,
		publicationActorID).Scan(&active, &row)
	if err != nil || active != uuid.MustParse(publicationOptionalV2ID) ||
		row != uuid.MustParse(publicationSourceID) {
		t.Fatalf("optional publication lineage = %s/%s/%v", active, row, err)
	}
}

func setPublicationJobInsert(
	t *testing.T, ctx context.Context, db publicationDatabase, allowed bool,
) {
	t.Helper()
	verb := "REVOKE INSERT ON platform.jobs FROM harpy_runtime"
	if allowed {
		verb = "GRANT INSERT ON platform.jobs TO harpy_runtime"
	}
	if _, err := db.admin.Exec(ctx, verb); err != nil {
		t.Fatalf("set publication enqueue permission: %v", err)
	}
}

func assertPublicationRolledBack(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	var plans, targets, jobs, receipts int
	err := db.admin.QueryRow(ctx, `SELECT
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_plans),
(SELECT count(*) FROM t_profile_publication.iam_profile_schema_versions
 WHERE id=$1),(SELECT count(*) FROM platform.jobs WHERE kind='iam.profile.backfill'),
(SELECT count(*) FROM t_profile_publication.iam_authorization_operation_receipts
 WHERE command_key='iam.profile_backfill.plan_create')`, publicationTargetV3ID).
		Scan(&plans, &targets, &jobs, &receipts)
	if err != nil || plans != 0 || targets != 0 || jobs != 0 || receipts != 0 {
		t.Fatalf("enqueue rollback = %d/%d/%d/%d/%v",
			plans, targets, jobs, receipts, err)
	}
}

func assertPublicationPlanAndFirstJob(
	t *testing.T, ctx context.Context, db publicationDatabase, planID uuid.UUID,
) {
	t.Helper()
	var source, target uuid.UUID
	var state, targetState, reference string
	var sourceCount, plans, jobs, mappings, normalized, retired, active int
	err := db.admin.QueryRow(ctx, `SELECT plan.source_schema_id,plan.target_schema_id,
plan.state,plan.source_count,target.state,page.payload_reference,
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_plans WHERE id=$1),
(SELECT count(*) FROM platform.jobs WHERE kind='iam.profile.backfill'),
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_mappings WHERE plan_id=$1),
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_mappings
 WHERE plan_id=$1 AND source_field_id=$2 AND target_field_id=$3
 AND converter_kind='copy_normalize_v1'),
(SELECT count(*) FROM t_profile_publication.iam_profile_field_revisions
 WHERE schema_version_id=plan.target_schema_id AND field_id=$2 AND field_state='retired'),
(SELECT count(*) FROM t_profile_publication.iam_profile_field_revisions
 WHERE schema_version_id=plan.target_schema_id AND field_id=$3 AND field_state='active')
FROM t_profile_publication.iam_profile_backfill_plans AS plan
JOIN t_profile_publication.iam_profile_schema_versions AS target
 ON target.id=plan.target_schema_id
JOIN t_profile_publication.iam_profile_backfill_pages AS page
 ON page.plan_id=plan.id AND page.page_number=1 WHERE plan.id=$1`, planID,
		publicationSourceField, publicationReplacement).
		Scan(&source, &target, &state, &sourceCount, &targetState, &reference,
			&plans, &jobs, &mappings, &normalized, &retired, &active)
	if err != nil || source != uuid.MustParse(publicationOptionalV2ID) ||
		target != uuid.MustParse(publicationTargetV3ID) || state != "planned" ||
		targetState != "draft" || sourceCount != 1 || plans != 1 || jobs != 1 ||
		mappings != 2 || normalized != 1 || retired != 1 || active != 1 ||
		!strings.Contains(reference, planID.String()) {
		t.Fatalf("plan and first job = %s/%s/%s/%d/%s/%q/%d/%d/%d/%d/%d/%d/%v",
			source, target, state, sourceCount, targetState, reference, plans, jobs,
			mappings, normalized, retired, active, err)
	}
}
