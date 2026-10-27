//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertRecoveryUserRootFence(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
) {
	t.Helper()
	rootID := uuid.MustParse("52000000-0000-4000-8000-000000000001")
	operationID := uuid.MustParse("53000000-0000-4000-8000-000000000001")
	plan := recoveryRootPlan(
		t, rootID, scenario.userV2, scenario.userV3,
		recoveryMappingsV3(), scenario.userRows, 2,
	)
	createRecoveryRootPlan(
		t, ctx, scenario.db, scenario.runtime, plan, operationID,
	)
	fenceRecoveryRootPlan(
		t, ctx, scenario.db, rootID, scenario.userV3.ID(),
		len(scenario.userRows), 2,
	)
	before := readRecoveryEvidence(t, ctx, scenario.db, rootID)
	assertRecoveryMixedRead(
		t, ctx, scenario.db, scenario.runtime,
		scenario.userV2.ID(), scenario.userV3.ID(),
	)
	assertRecoveryRawFenceState(t, ctx, scenario.db)
	assertRecoveryFencedWriteError(
		t, ctx, scenario.db, scenario.runtime,
	)
	assertRecoveryRuntimeStages(t, ctx, scenario, rootID)
	assertRecoveryCursorEvidenceCAS(t, ctx, scenario, rootID)
	childID := assertRecoverySameKeyConcurrency(t, ctx, scenario, rootID)
	assertRecoveryLineageCounts(
		t, ctx, scenario.db, rootID,
		map[int64]string{0: "fenced", 1: "planned"},
	)
	assertRecoveryEvidenceUnchanged(t, ctx, scenario.db, rootID, before)
	assertRecoveryNonzeroActivation(t, ctx, scenario, rootID, childID)
	assertRecoveryEvidenceUnchanged(t, ctx, scenario.db, rootID, before)
	assertRecoveryLinearChildAndZero(t, ctx, scenario)
}
