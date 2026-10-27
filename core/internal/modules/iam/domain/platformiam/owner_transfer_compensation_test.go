package platformiam

import (
	"testing"
	"time"
)

func TestOwnerTransferCompensationRequiresHigherGeneration(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	command := OwnerTransferCompensation{
		Evidence: fixtureEvidence(), OriginalOperationID: fixtureID(9),
		OriginalAckSHA256: [32]byte{1}, ProposalSHA256: [32]byte{2},
		ProposalSignerKID: "registry-1", ProposalSignatureSHA256: [32]byte{3},
		ProposalExpiresAt: now.Add(10 * time.Minute), ExpectedOriginalGeneration: 2,
	}
	command.StepUp = fixtureStepUp(now, OwnerTransferCompensationPurpose(command))
	if err := command.Validate(now); err != nil {
		t.Fatalf("valid compensation rejected: %v", err)
	}
	command.ExpectedOriginalGeneration = 1
	if err := command.Validate(now); err == nil {
		t.Fatal("compensation accepted a non-transfer generation")
	}
}
