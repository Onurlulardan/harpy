//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func fenceRecoveryRootPlan(
	t *testing.T, ctx context.Context, db publicationDatabase,
	planID uuid.UUID, targetID uuid.UUID, totalRows, fencePage int,
) {
	t.Helper()
	for page := 1; page <= fencePage; page++ {
		outcome := executeRecoveryWorkerPage(t, ctx, db, planID, page)
		if page < fencePage && outcome.Successor == nil {
			t.Fatalf("recovery root page %d has no successor", page)
		}
		if page == fencePage && outcome.Successor != nil {
			t.Fatalf("fenced recovery root page %d has successor %#v", page, outcome.Successor)
		}
	}
	var state, pageState string
	var completed, converted, failures, targets, remaining int
	err := db.admin.QueryRow(ctx, `SELECT plan.state,page.state,
checkpoint.completed_pages,checkpoint.converted_count,checkpoint.error_count,
(SELECT count(*) FROM t_profile_publication.iam_profile_recovery_resources_v1
 WHERE schema_version_id=$2),
(SELECT count(*) FROM t_profile_publication.iam_profile_recovery_resources_v1
 WHERE resource_kind=plan.resource_kind AND schema_version_id<>$2)
FROM t_profile_publication.iam_profile_backfill_plans AS plan
JOIN t_profile_publication.iam_profile_backfill_pages AS page
 ON page.plan_id=plan.id AND page.page_number=$3
JOIN t_profile_publication.iam_profile_backfill_checkpoints AS checkpoint
 ON checkpoint.plan_id=plan.id WHERE plan.id=$1`, planID, targetID, fencePage).Scan(
		&state, &pageState, &completed, &converted, &failures, &targets, &remaining,
	)
	if err != nil || state != "fenced" || pageState != "fenced" ||
		completed != fencePage || converted != fencePage || failures != 1 ||
		targets != fencePage || remaining != totalRows-fencePage {
		t.Fatalf("recovery root fence = %s/%s/%d/%d/%d/%d/%d/%v",
			state, pageState, completed, converted, failures, targets, remaining, err)
	}
}

func assertRecoveryLineageCounts(
	t *testing.T, ctx context.Context, db publicationDatabase,
	rootID uuid.UUID, states map[int64]string,
) {
	t.Helper()
	rows, err := db.admin.Query(ctx, `SELECT COALESCE(recovery_sequence,0),state
FROM t_profile_publication.iam_profile_backfill_plans
WHERE id=$1 OR recovery_root_plan_id=$1 ORDER BY COALESCE(recovery_sequence,0)`, rootID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := make(map[int64]string, len(states))
	for rows.Next() {
		var sequence int64
		var state string
		if err = rows.Scan(&sequence, &state); err != nil {
			t.Fatal(err)
		}
		seen[sequence] = state
	}
	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}
	if len(seen) != len(states) {
		t.Fatalf("recovery lineage states = %#v, want %#v", seen, states)
	}
	for sequence, state := range states {
		if seen[sequence] != state {
			t.Fatalf("recovery lineage sequence %d = %q, want %q", sequence, seen[sequence], state)
		}
	}
}
