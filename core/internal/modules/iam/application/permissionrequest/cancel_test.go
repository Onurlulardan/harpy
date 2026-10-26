package permissionrequest

import (
	"testing"

	"github.com/google/uuid"
)

func TestCancelCommandRequiresMutableStateAndCurrentVersion(t *testing.T) {
	command := CancelCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		GroupID: uuid.New(), ExpectedRevision: 1, ExpectedState: StateInReview,
		RequesterID: uuid.New(), ExpectedAuthorizationVersion: 1,
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid cancel command: %v", err)
	}
	command.ExpectedState = StateApproved
	if err := command.Validate(); err == nil {
		t.Fatal("terminal request cancellation was accepted")
	}
}
