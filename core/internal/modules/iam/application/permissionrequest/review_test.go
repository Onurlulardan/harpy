package permissionrequest

import (
	"testing"

	"github.com/google/uuid"
)

func TestReviewCommandRequiresCurrentReviewerVersion(t *testing.T) {
	command := ReviewCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		GroupID: uuid.New(), ExpectedRevision: 1, ExpectedState: StatePending,
		ReviewerID: uuid.New(), ExpectedReviewerAuthorizationVersion: 1,
		ReviewTaskID: uuid.New(),
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid review command: %v", err)
	}
	command.ExpectedReviewerAuthorizationVersion = 0
	if err := command.Validate(); err == nil {
		t.Fatal("zero reviewer authorization version was accepted")
	}
}
