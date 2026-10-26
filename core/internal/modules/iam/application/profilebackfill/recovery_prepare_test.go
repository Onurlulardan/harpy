package profilebackfill

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestPrepareRecoveryEncodesZeroRemainingWithoutJob(t *testing.T) {
	parent := recoveryFixtureParent()
	plan := recoveryFixturePlan(t, parent, 0)
	prepared, err := PrepareRecovery(RecoveryCreateCommand{
		Plan: plan, Evidence: recoveryFixtureEvidence(),
	}, uuid.MustParse("17000000-0000-4000-8000-000000000001"))
	if err != nil {
		t.Fatal(err)
	}
	if prepared.FirstJob != nil || prepared.PlanSHA256 != sha256.Sum256(prepared.Document) {
		t.Fatalf("zero recovery job/digest = %+v/%x", prepared.FirstJob, prepared.PlanSHA256)
	}
	var wire recoveryPlanWire
	if err := json.Unmarshal(prepared.Document, &wire); err != nil {
		t.Fatal(err)
	}
	wantEmptySHA := sha256.Sum256(nil)
	if wire.RecoveryOfPlanID != parent.ID.String() ||
		wire.RecoveryRootPlanID != parent.ID.String() || wire.RecoverySequence != 1 ||
		wire.ParentPlanSHA256 != hex.EncodeToString(parent.PlanSHA256[:]) ||
		wire.SourceSchemaID != parent.SourceSchemaID.String() ||
		wire.TargetSchemaID != parent.Target.ID.String() ||
		wire.InheritedTargetCount != 0 || wire.RemainingSourceCount != 0 ||
		wire.InheritedTargetSHA256 != hex.EncodeToString(wantEmptySHA[:]) ||
		wire.RemainingSourceSHA256 != hex.EncodeToString(wantEmptySHA[:]) ||
		wire.ExpectedRemainingTargetSHA256 != hex.EncodeToString(wantEmptySHA[:]) ||
		len(wire.Mappings) != 1 || len(wire.Pages) != 0 {
		t.Fatalf("unexpected recovery wire: %+v", wire)
	}
}

func TestPrepareRecoveryReturnsExistingWorkerContract(t *testing.T) {
	parent := recoveryFixtureParent()
	plan := recoveryFixturePlan(t, parent, 1)
	tenantID := uuid.MustParse("17000000-0000-4000-8000-000000000001")
	prepared, err := PrepareRecovery(RecoveryCreateCommand{
		Plan: plan, Evidence: recoveryFixtureEvidence(),
	}, tenantID)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.FirstJob == nil || prepared.FirstJob.PlanID != plan.ID() ||
		prepared.FirstJob.PayloadReference !=
			"urn:harpy:iam:profile-backfill:"+tenantID.String()+":"+
				plan.ID().String()+":1:v1" {
		t.Fatalf("first job = %+v", prepared.FirstJob)
	}
}
