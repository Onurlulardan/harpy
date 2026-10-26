package permissionrequest

import (
	"testing"

	"github.com/google/uuid"
)

func TestExpireCommandAllowsOnlyPendingOrReview(t *testing.T) {
	command := ExpireCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		GroupID: uuid.New(), ExpectedRevision: 1, ExpectedState: StatePending,
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid expire command: %v", err)
	}
	command.ExpectedState = StateRejected
	if err := command.Validate(); err == nil {
		t.Fatal("terminal request expiry was accepted")
	}
}
