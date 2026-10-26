//go:build integration

package integrationtest

import (
	"context"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func assertRecoveryZeroActivation(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
	rootID, parentID, sourceID, targetID uuid.UUID,
) uuid.UUID {
	t.Helper()
	input := iammodule.ProfileBackfillRecoveryInput{
		IdempotencyKey: "recovery-user-v4-zero",
		Resource:       profileapp.ResourceUser, ParentPlanID: parentID,
	}
	claims := publicationClaims(t, scenario.db.tenant)
	created, err := scenario.runtime.application.RecoverBackfill(ctx, claims, input)
	if err != nil || created.PlanID == uuid.Nil || created.RecoveryOfPlanID != parentID ||
		created.RecoveryRootPlanID != rootID || created.RecoverySequence != 2 ||
		created.State != profilebackfillapp.RecoveryActivated || created.Replayed {
		t.Fatalf("zero recovery creation = %#v/%v", created, err)
	}
	replay, err := scenario.runtime.application.RecoverBackfill(ctx, claims, input)
	if err != nil || replay.PlanID != created.PlanID || !replay.Replayed ||
		replay.State != profilebackfillapp.RecoveryActivated ||
		replay.Revision != created.Revision ||
		replay.AuthorizationClock != created.AuthorizationClock {
		t.Fatalf("zero recovery replay = %#v/%v", replay, err)
	}
	assertRecoveryCreateCounts(t, ctx, scenario.db, parentID, created.PlanID, 0)
	assertRecoveryLineageCounts(t, ctx, scenario.db, rootID,
		map[int64]string{0: "fenced", 1: "fenced", 2: "activated"})
	var sourceState, targetState string
	err = scenario.db.admin.QueryRow(ctx, `SELECT source.state,target.state
FROM t_profile_publication.iam_profile_schema_versions AS source
JOIN t_profile_publication.iam_profile_schema_versions AS target ON target.id=$2
WHERE source.id=$1`, sourceID, targetID).Scan(&sourceState, &targetState)
	if err != nil || sourceState != "retired" || targetState != "active" {
		t.Fatalf("zero recovery schema activation = %s/%s/%v",
			sourceState, targetState, err)
	}
	return created.PlanID
}
