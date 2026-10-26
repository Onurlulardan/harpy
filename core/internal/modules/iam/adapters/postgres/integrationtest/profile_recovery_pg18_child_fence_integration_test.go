//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertRecoveryDriftChildFences(
	t *testing.T, ctx context.Context, db publicationDatabase,
	rootID, childID, targetID uuid.UUID,
) {
	t.Helper()
	outcome := executeRecoveryWorkerPage(t, ctx, db, childID, 1)
	if outcome.Successor != nil {
		t.Fatalf("drift child successor = %#v", outcome.Successor)
	}
	var state, pageState string
	var completed, converted, failures, targets, remaining int
	err := db.admin.QueryRow(ctx, `SELECT plan.state,page.state,
checkpoint.completed_pages,checkpoint.converted_count,checkpoint.error_count,
(SELECT count(*) FROM t_profile_publication.iam_profile_recovery_resources_v1
 WHERE resource_kind='user' AND schema_version_id=$2),
(SELECT count(*) FROM t_profile_publication.iam_profile_recovery_resources_v1
 WHERE resource_kind='user' AND schema_version_id<>$2)
FROM t_profile_publication.iam_profile_backfill_plans AS plan
JOIN t_profile_publication.iam_profile_backfill_pages AS page
 ON page.plan_id=plan.id AND page.page_number=1
JOIN t_profile_publication.iam_profile_backfill_checkpoints AS checkpoint
 ON checkpoint.plan_id=plan.id WHERE plan.id=$1`, childID, targetID).Scan(
		&state, &pageState, &completed, &converted, &failures, &targets, &remaining,
	)
	if err != nil || state != "fenced" || pageState != "fenced" ||
		completed != 1 || converted != 1 || failures != 1 || targets != 3 ||
		remaining != 0 {
		t.Fatalf("drift recovery child fence = %s/%s/%d/%d/%d/%d/%d/%v",
			state, pageState, completed, converted, failures, targets, remaining, err)
	}
	assertRecoveryLineageCounts(t, ctx, db, rootID,
		map[int64]string{0: "fenced", 1: "fenced"})
}
