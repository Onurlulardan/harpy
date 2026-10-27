package restoreexecution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

var fixtureIdentity = ports.RestoreExecutionIdentity{
	OperationID: "11111111-1111-4111-8111-111111111111",
	AttemptID:   "22222222-2222-7222-8222-222222222222",
	LeaseToken:  "33333333-3333-7333-8333-333333333333",
}

type storeFixture struct {
	claim   ports.RestoreExecutionClaim
	found   bool
	changed bool
}

func (fixture *storeFixture) SeedRehydrateControl(
	_ context.Context,
	_ ports.RehydrateControlSeed,
) (bool, error) {
	return true, nil
}

func (fixture *storeFixture) ClaimRestoreExecution(
	context.Context, ports.RestoreExecutionIdentity, string,
) (ports.RestoreExecutionClaim, bool, error) {
	return fixture.claim, fixture.found, nil
}
func (fixture *storeFixture) UpdateRestoreLease(
	context.Context, ports.RestoreExecutionIdentity,
) (bool, error) {
	return fixture.changed, nil
}
func (fixture *storeFixture) StartRestoreDrain(
	context.Context, ports.RestoreExecutionIdentity,
) (bool, error) {
	return fixture.changed, nil
}
func (fixture *storeFixture) ExecuteRestoreApply(
	context.Context, ports.RestoreApplyEvidence,
) (bool, error) {
	return fixture.changed, nil
}
func (fixture *storeFixture) ValidateRestoreExecution(
	context.Context, ports.RestoreVerification,
) (bool, error) {
	return fixture.changed, nil
}
func (fixture *storeFixture) MarkRestoreSwitched(
	context.Context, ports.RestoreSwitchEvidence,
) (bool, error) {
	return fixture.changed, nil
}
func (fixture *storeFixture) CompleteRestoreExecution(
	context.Context, ports.RestoreCompletion,
) (bool, error) {
	return fixture.changed, nil
}
func (fixture *storeFixture) RejectRestoreExecution(
	context.Context, ports.RestoreFailure,
) (bool, error) {
	return fixture.changed, nil
}
func (fixture *storeFixture) ImportRestoreReceipt(
	context.Context, ports.RestoreReceiptImport,
) (bool, error) {
	return fixture.changed, nil
}

func validClaimFixture() ports.RestoreExecutionClaim {
	plan, manifest := []byte("signed-plan"), []byte("signed-manifest")
	planDigest, manifestDigest := sha256.Sum256(plan), sha256.Sum256(manifest)
	return ports.RestoreExecutionClaim{
		RestoreExecutionIdentity: fixtureIdentity,
		AttemptGeneration:        1,
		RestoreState:             "plan_ready",
		RecoverySetID:            "44444444-4444-4444-8444-444444444444",
		EnvironmentID:            "55555555-5555-4555-8555-555555555555",
		TargetDisposition:        "existing_readable",
		PlanSHA256:               hex.EncodeToString(planDigest[:]),
		PlanBytes:                plan,
		SafetyRecoverySetID:      "66666666-6666-4666-8666-666666666666",
		SafetyRecoverySetSHA256:  digestFixture(),
		ManifestSHA256:           hex.EncodeToString(manifestDigest[:]),
		ManifestBytes:            manifest,
		PreviousDatabaseSystemID: "77777777-7777-4777-8777-777777777777",
		ComponentInventory:       componentFixture(),
		DatabaseNow:              time.Date(2026, 7, 23, 1, 2, 3, 0, time.UTC),
	}
}

func componentFixture() json.RawMessage {
	values := map[string]any{}
	for name := range requiredComponents {
		values[name] = componentInventoryItem{
			ObjectKey: "objects/" + name, ObjectVersionID: "version-" + name,
			CiphertextSHA256: digestFixture(), CiphertextBytes: 1,
		}
	}
	document, _ := json.Marshal(values)
	return document
}

func digestFixture() string {
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
