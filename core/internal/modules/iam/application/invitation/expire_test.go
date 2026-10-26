package invitation

import (
	"testing"

	"github.com/google/uuid"
)

func TestExpireCommandRequiresCASAndCurrentActorVersion(t *testing.T) {
	command := ExpireCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		ID: uuid.New(), ExpectedRevision: 1, ActorID: uuid.New(),
		ExpectedAuthorizationVersion: 1,
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid expire command: %v", err)
	}
	command.ExpectedAuthorizationVersion = 0
	if err := command.Validate(); err == nil {
		t.Fatal("zero actor authorization version was accepted")
	}
}
