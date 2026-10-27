package invitation

import (
	"testing"

	"github.com/google/uuid"
)

func TestRevokeCommandRequiresCASAndCurrentActorVersion(t *testing.T) {
	command := RevokeCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		ID: uuid.New(), ExpectedRevision: 1, ActorID: uuid.New(),
		ExpectedAuthorizationVersion: 1,
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid revoke command: %v", err)
	}
	command.ExpectedRevision = 0
	if err := command.Validate(); err == nil {
		t.Fatal("zero invitation revision was accepted")
	}
}
