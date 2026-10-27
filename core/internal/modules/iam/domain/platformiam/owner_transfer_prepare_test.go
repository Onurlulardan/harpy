package platformiam

import (
	"testing"
	"time"
)

func TestOwnerTransferPrepareRequiresDistinctMonotonicHead(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	command := OwnerTransferPrepare{
		Evidence: fixtureEvidence(), TargetUserID: fixtureID(3),
		ActorExpectedRevision: 6, TargetExpectedRevision: 7,
		TargetExpectedTokenVersion: 8, TargetExpectedAuthorizationVersion: 9,
		SuccessorSubjectID: fixtureID(4), SuccessorBindingID: fixtureID(5),
		RegistryBindingID: fixtureID(6), ExpectedOwnerPolicyRevision: 2,
		ExpectedHeadGeneration: 1, ExpectedHeadSHA256: [32]byte{1},
		StateSHA256: [32]byte{2}, ProposalSHA256: [32]byte{3},
		ProposalSignerKID: "registry-1", ProposalSignatureSHA256: [32]byte{4},
		ProposalExpiresAt: now.Add(10 * time.Minute),
	}
	command.StepUp = fixtureStepUp(now, OwnerTransferPurpose(command))
	if err := command.Validate(now); err != nil {
		t.Fatalf("valid owner transfer rejected: %v", err)
	}
	command.StateSHA256 = command.ExpectedHeadSHA256
	if err := command.Validate(now); err == nil {
		t.Fatal("owner transfer accepted an unchanged protected head")
	}
}
