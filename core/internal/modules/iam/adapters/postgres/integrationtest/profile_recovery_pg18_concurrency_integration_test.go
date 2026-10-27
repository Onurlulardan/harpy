//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

type recoveryPublicCall struct {
	view iammodule.ProfileBackfillRecoveryView
	err  error
}

func assertRecoveryDifferentKeyConcurrency(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
	parentID uuid.UUID,
) uuid.UUID {
	t.Helper()
	raceCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	calls := recoverConcurrently(
		raceCtx, scenario.runtime, publicationClaims(t, scenario.db.tenant), parentID,
		profileapp.ResourceOrganization, "recovery-org-a", "recovery-org-b",
	)
	var winner *iammodule.ProfileBackfillRecoveryView
	conflicts := 0
	for index := range calls {
		call := calls[index]
		if call.err == nil {
			if winner != nil {
				t.Fatalf("different-key recovery created twice: %#v", calls)
			}
			winner = &call.view
			continue
		}
		if errors.Is(call.err, iammodule.ErrProfileHTTPConflict) {
			conflicts++
			continue
		}
		t.Fatalf("different-key recovery error = %v", call.err)
	}
	if winner == nil || conflicts != 1 || winner.PlanID == uuid.Nil ||
		winner.RecoveryOfPlanID != parentID ||
		winner.RecoveryRootPlanID != parentID || winner.RecoverySequence != 1 ||
		winner.State != profilebackfillapp.RecoveryPlanned || winner.Replayed {
		t.Fatalf("different-key recovery winner/conflicts = %#v/%d", winner, conflicts)
	}
	assertRecoveryCreateCounts(t, ctx, scenario.db, parentID, winner.PlanID, 1)
	return winner.PlanID
}

func recoverConcurrently(
	ctx context.Context, runtime publicationRuntime, claims accesstoken.Claims,
	parentID uuid.UUID, resource profileapp.ResourceKind, keys ...string,
) []recoveryPublicCall {
	start := make(chan struct{})
	results := make(chan recoveryPublicCall, len(keys))
	for _, key := range keys {
		key := key
		go func() {
			<-start
			view, err := runtime.application.RecoverBackfill(
				ctx, claims,
				iammodule.ProfileBackfillRecoveryInput{
					IdempotencyKey: key, Resource: resource, ParentPlanID: parentID,
				},
			)
			results <- recoveryPublicCall{view: view, err: err}
		}()
	}
	close(start)
	output := make([]recoveryPublicCall, 0, len(keys))
	for range keys {
		select {
		case result := <-results:
			output = append(output, result)
		case <-ctx.Done():
			output = append(output, recoveryPublicCall{err: ctx.Err()})
		}
	}
	return output
}

func assertRecoverySameKeyConcurrency(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
	parentID uuid.UUID,
) uuid.UUID {
	t.Helper()
	raceCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	calls := recoverConcurrently(
		raceCtx, scenario.runtime, publicationClaims(t, scenario.db.tenant), parentID,
		profileapp.ResourceUser, "recovery-user-v3", "recovery-user-v3",
	)
	if len(calls) != 2 || calls[0].err != nil || calls[1].err != nil {
		t.Fatalf("same-key recovery calls = %#v/%v and %#v/%v",
			calls[0].view, calls[0].err, calls[1].view, calls[1].err)
	}
	first, second := calls[0].view, calls[1].view
	if first.PlanID == uuid.Nil || first.PlanID != second.PlanID ||
		first.RecoveryOfPlanID != parentID || second.RecoveryOfPlanID != parentID ||
		first.RecoveryRootPlanID != parentID || second.RecoveryRootPlanID != parentID ||
		first.RecoverySequence != 1 || second.RecoverySequence != 1 ||
		first.State != profilebackfillapp.RecoveryPlanned ||
		second.State != profilebackfillapp.RecoveryPlanned ||
		first.Replayed == second.Replayed {
		t.Fatalf("same-key recovery results = %#v/%#v", first, second)
	}
	assertRecoveryCreateCounts(t, ctx, scenario.db, parentID, first.PlanID, 1)
	return first.PlanID
}

func assertRecoveryCreateCounts(
	t *testing.T, ctx context.Context, db publicationDatabase,
	parentID, childID uuid.UUID, expectedJobs int,
) {
	t.Helper()
	var children, receipts, pages, jobs int
	err := db.admin.QueryRow(ctx, `SELECT
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_plans
 WHERE recovery_of_plan_id=$1),
(SELECT count(*) FROM t_profile_publication.iam_authorization_operation_receipts
 WHERE command_key='iam.profile_backfill.recovery_create' AND result_id=$2),
(SELECT count(*) FROM t_profile_publication.iam_profile_backfill_pages
 WHERE plan_id=$2),
(SELECT count(*) FROM platform.jobs AS job JOIN
 t_profile_publication.iam_profile_backfill_pages AS page ON page.job_id=job.id
 WHERE page.plan_id=$2)`, parentID, childID).Scan(
		&children, &receipts, &pages, &jobs,
	)
	if err != nil || children != 1 || receipts != 1 || pages != expectedJobs ||
		jobs != expectedJobs {
		t.Fatalf("recovery create counts = %d/%d/%d/%d/%v",
			children, receipts, pages, jobs, err)
	}
}
