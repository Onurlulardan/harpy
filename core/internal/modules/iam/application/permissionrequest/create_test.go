package permissionrequest

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateCommandRejectsStaleAndCrossTargetSelfService(t *testing.T) {
	command := validCreateCommand()
	if err := command.Validate(); err != nil {
		t.Fatalf("valid create command: %v", err)
	}
	command.ExpectedAuthorizationVersion = 0
	if err := command.Validate(); err == nil {
		t.Fatal("zero authorization version was accepted")
	}
	command = validCreateCommand()
	command.Requests[0].Target.ID = uuid.New()
	if err := command.Validate(); err != ErrDenied {
		t.Fatalf("cross-target self-service error = %v", err)
	}
}
