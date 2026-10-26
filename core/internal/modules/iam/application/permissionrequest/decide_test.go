package permissionrequest

import (
	"testing"

	"github.com/google/uuid"
)

func TestDecideCommandRequiresPurposeBoundStepUpAndCurrentVersions(t *testing.T) {
	command := DecideCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1}, GroupID: uuid.New(),
		ExpectedRevision: 2, ExpectedState: StateInReview,
		ReviewerID: uuid.New(), ExpectedReviewerTokenVersion: 1,
		ExpectedReviewerAuthorizationVersion: 1, ReviewTaskID: uuid.New(),
		Decision: DecisionApprove, StepUp: validStepUpEvidence(),
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid decision command: %v", err)
	}
	command.StepUp.ACR = "step_up"
	if err := command.Validate(); err == nil {
		t.Fatal("noncanonical step-up ACR was accepted")
	}
	command.StepUp = validStepUpEvidence()
	command.ExpectedReviewerAuthorizationVersion = 0
	if err := command.Validate(); err == nil {
		t.Fatal("zero reviewer authorization version was accepted")
	}
}
