package profilebackfill

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRecoveryPlanBuildsLinearChainAndRejectsFork(t *testing.T) {
	root := recoveryFixtureParent()
	first := recoveryFixturePlan(t, root, 1)
	if first.RecoveryOfPlanID() != root.ID ||
		first.RecoveryRootPlanID() != root.ID || first.RecoverySequence() != 1 {
		t.Fatalf("first lineage = %s/%s/%d", first.RecoveryOfPlanID(),
			first.RecoveryRootPlanID(), first.RecoverySequence())
	}
	parent := RecoveryParent{
		ID: first.ID(), RecoveryRootPlanID: first.RecoveryRootPlanID(),
		RecoverySequence: first.RecoverySequence(), State: RecoveryParentFenced,
		PlanSHA256:     sha256.Sum256([]byte("first")),
		SourceSchemaID: first.SourceSchemaID(), Target: first.Target(),
		ConverterVersion: first.ConverterVersion(), Mappings: first.Mappings(),
	}
	secondID := uuid.MustParse("15000000-0000-4000-8000-000000000003")
	second, err := PlanRecovery(
		secondID, parent, recoveryFixtureInventory(t, parent, 1), MaximumPageSize,
		time.Date(2026, 7, 16, 13, 0, 0, 0, time.UTC),
	)
	if err != nil || second.RecoveryOfPlanID() != first.ID() ||
		second.RecoveryRootPlanID() != root.ID || second.RecoverySequence() != 2 {
		t.Fatalf("second lineage = %+v/%v", second, err)
	}
	parent.ExistingChildPlanID = second.ID()
	_, err = PlanRecovery(
		uuid.MustParse("15000000-0000-4000-8000-000000000004"), parent,
		recoveryFixtureInventory(t, parent, 1), MaximumPageSize,
		time.Date(2026, 7, 16, 14, 0, 0, 0, time.UTC),
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("fork error = %v", err)
	}
}

func TestNormalPlanStillRejectsZeroPages(t *testing.T) {
	parent := recoveryFixtureParent()
	plan := Plan{
		ID:             uuid.MustParse("15000000-0000-4000-8000-000000000005"),
		SourceSchemaID: parent.SourceSchemaID, Target: parent.Target,
		ConverterVersion: ConverterTypedMap, SourceCount: 0,
		PageSize: MaximumPageSize, InitialRollingSHA256: sha256.Sum256(nil),
		Mappings: parent.Mappings, Pages: []Page{},
	}
	if !errors.Is(plan.Validate(), ErrInvalid) {
		t.Fatal("normal zero-page plan passed")
	}
}
