//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertRecoveryLinearChildAndZero(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
) {
	t.Helper()
	target := recoveryUserTargetV4(t)
	rows := readRecoveryUserRows(t, ctx, scenario.db, scenario.userV2, scenario.userV3)
	rootID := uuid.MustParse("52000000-0000-4000-8000-000000000002")
	rootOperation := uuid.MustParse("53000000-0000-4000-8000-000000000002")
	plan := recoveryRootPlan(
		t, rootID, scenario.userV3, target, recoveryMappingsV4(), rows, 2,
	)
	createRecoveryRootPlan(
		t, ctx, scenario.db, scenario.runtime, plan, rootOperation,
	)
	fenceRecoveryRootPlan(
		t, ctx, scenario.db, rootID, target.ID(), len(rows), 2,
	)
	rootEvidence := readRecoveryEvidence(t, ctx, scenario.db, rootID)
	childID := uuid.MustParse("54000000-0000-4000-8000-000000000002")
	childOperation := uuid.MustParse("55000000-0000-4000-8000-000000000002")
	createRecoveryDriftChild(
		t, ctx, scenario, rootID, childID, childOperation,
	)
	assertRecoveryCreateCounts(t, ctx, scenario.db, rootID, childID, 1)
	assertRecoveryDriftChildFences(
		t, ctx, scenario.db, rootID, childID, target.ID(),
	)
	childEvidence := readRecoveryEvidence(t, ctx, scenario.db, childID)
	assertRecoveryZeroActivation(
		t, ctx, scenario, rootID, childID, scenario.userV3.ID(), target.ID(),
	)
	assertRecoveryEvidenceUnchanged(t, ctx, scenario.db, rootID, rootEvidence)
	assertRecoveryEvidenceUnchanged(t, ctx, scenario.db, childID, childEvidence)
}
