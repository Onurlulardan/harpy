//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

type recoveryWriteCounts struct {
	plans, children, receipts, jobs int
}

func assertRecoveryFutureRejectedBeforeWrites(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
) {
	t.Helper()
	source := recoveryUserTargetV4(t)
	target := recoveryUserTargetV5(t)
	rows := readRecoveryUserRows(t, ctx, scenario.db, source)
	rootID := uuid.MustParse("52000000-0000-4000-8000-000000000003")
	operationID := uuid.MustParse("53000000-0000-4000-8000-000000000003")
	plan := recoveryRootPlan(
		t, rootID, source, target, recoveryMappingsV5(), rows, 1,
	)
	createRecoveryRootPlan(
		t, ctx, scenario.db, scenario.runtime, plan, operationID,
	)
	fenceRecoveryRootPlan(
		t, ctx, scenario.db, rootID, target.ID(), len(rows), 1,
	)
	seedRecoveryFutureUser(t, ctx, scenario.db)
	rootEvidence := readRecoveryEvidence(t, ctx, scenario.db, rootID)
	before := readRecoveryWriteCounts(t, ctx, scenario.db, rootID)
	view, err := scenario.runtime.application.RecoverBackfill(
		ctx, publicationClaims(t, scenario.db.tenant),
		iammodule.ProfileBackfillRecoveryInput{
			IdempotencyKey: "recovery-future-reject",
			Resource:       profileapp.ResourceUser, ParentPlanID: rootID,
		},
	)
	if !errors.Is(err, iammodule.ErrProfileHTTPUnavailable) ||
		view != (iammodule.ProfileBackfillRecoveryView{}) {
		t.Fatalf("future recovery rejection = %#v/%v", view, err)
	}
	after := readRecoveryWriteCounts(t, ctx, scenario.db, rootID)
	if after != before {
		t.Fatalf("future recovery wrote state = %#v -> %#v", before, after)
	}
	assertRecoveryEvidenceUnchanged(t, ctx, scenario.db, rootID, rootEvidence)
	assertRecoveryLineageCounts(t, ctx, scenario.db, rootID,
		map[int64]string{0: "fenced"})
}

func readRecoveryWriteCounts(
	t *testing.T, ctx context.Context, db publicationDatabase, rootID uuid.UUID,
) recoveryWriteCounts {
	t.Helper()
	var result recoveryWriteCounts
	err := db.admin.QueryRow(ctx, `SELECT
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_plans),
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_plans
 WHERE recovery_of_plan_id=$1),
(SELECT count(*) FROM t_profile_publication.iam_authorization_operation_receipts
 WHERE command_key='iam.profile_backfill.recovery_create'),
(SELECT count(*) FROM platform.jobs)`, rootID).Scan(
		&result.plans, &result.children, &result.receipts, &result.jobs,
	)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
