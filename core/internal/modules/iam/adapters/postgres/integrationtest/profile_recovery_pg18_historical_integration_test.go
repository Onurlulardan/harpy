//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertRecoveryHistoricalEmptyMapping(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
) {
	t.Helper()
	source := publicationEmptyCurrentSchema(t)
	target := recoveryOrganizationTargetV3(t)
	rows := seedRecoveryOrganizationRow(t, ctx, scenario.db)
	rootID := uuid.MustParse("52000000-0000-4000-8000-000000000101")
	operationID := uuid.MustParse("53000000-0000-4000-8000-000000000101")
	plan := recoveryRootPlan(
		t, rootID, source, target, recoveryOrganizationMappings(), rows, 1,
	)
	createRecoveryRootPlan(
		t, ctx, scenario.db, scenario.runtime, plan, operationID,
	)
	fenceRecoveryRootPlan(
		t, ctx, scenario.db, rootID, target.ID(), len(rows), 1,
	)
	rootEvidence := readRecoveryEvidence(t, ctx, scenario.db, rootID)
	childID := assertRecoveryDifferentKeyConcurrency(t, ctx, scenario, rootID)
	assertRecoveryHistoricalPage(t, ctx, scenario.db, childID)
	executeRecoveryWorkerPage(t, ctx, scenario.db, childID, 1)
	assertRecoveryHistoricalActivation(
		t, ctx, scenario.db, rootID, childID, source.ID(), target.ID(),
	)
	assertRecoveryEvidenceUnchanged(t, ctx, scenario.db, rootID, rootEvidence)
}

func assertRecoveryHistoricalPage(
	t *testing.T, ctx context.Context, db publicationDatabase, childID uuid.UUID,
) {
	t.Helper()
	var after, through uuid.UUID
	var expected, inherited, remaining int
	err := db.admin.QueryRow(ctx, `SELECT page.after_resource_id,page.through_resource_id,
page.expected_count,plan.inherited_target_count,plan.remaining_source_count
FROM t_profile_publication.iam_profile_backfill_pages AS page
JOIN t_profile_publication.iam_profile_backfill_plans AS plan ON plan.id=page.plan_id
WHERE page.plan_id=$1 AND page.page_number=1`, childID).Scan(
		&after, &through, &expected, &inherited, &remaining,
	)
	if err != nil || after != uuid.Nil ||
		through != uuid.MustParse(publicationEmptyOrganization) || expected != 1 ||
		inherited != 1 || remaining != 1 {
		t.Fatalf("historical recovery page = %s/%s/%d/%d/%d/%v",
			after, through, expected, inherited, remaining, err)
	}
}

func assertRecoveryHistoricalActivation(
	t *testing.T, ctx context.Context, db publicationDatabase,
	rootID, childID, sourceID, targetID uuid.UUID,
) {
	t.Helper()
	var state, sourceState, targetState, historical string
	var historicalSchema uuid.UUID
	err := db.admin.QueryRow(ctx, `SELECT plan.state,source.state,target.state,
historical.profile_schema_version_id,historical.profile_values::text
FROM t_profile_publication.iam_profile_backfill_plans AS plan
JOIN t_profile_publication.iam_profile_schema_versions AS source ON source.id=$2
JOIN t_profile_publication.iam_profile_schema_versions AS target ON target.id=$3
JOIN t_profile_publication.iam_organizations AS historical ON historical.id=$4
WHERE plan.id=$1`, childID, sourceID, targetID, publicationEmptyOrganization).Scan(
		&state, &sourceState, &targetState, &historicalSchema, &historical,
	)
	if err != nil || state != "activated" || sourceState != "retired" ||
		targetState != "active" || historicalSchema != targetID || historical != "{}" {
		t.Fatalf("historical recovery activation = %s/%s/%s/%s/%s/%v",
			state, sourceState, targetState, historicalSchema, historical, err)
	}
	assertRecoveryLineageCounts(t, ctx, db, rootID,
		map[int64]string{0: "fenced", 1: "activated"})
}
