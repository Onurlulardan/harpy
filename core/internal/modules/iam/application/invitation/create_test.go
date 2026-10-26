package invitation

import "testing"

func TestCreateCommandRequiresVerifiedShapeAndCurrentVersion(t *testing.T) {
	command := validCreateCommand()
	if err := command.Validate(); err != nil {
		t.Fatalf("valid create command: %v", err)
	}
	command.ExpectedAuthorizationVersion = 0
	if err := command.Validate(); err == nil {
		t.Fatal("zero inviter authorization version was accepted")
	}
	command = validCreateCommand()
	command.Token.DeliveryArtifactRef = "plaintext-token"
	if err := command.Validate(); err == nil {
		t.Fatal("unsafe delivery artifact reference was accepted")
	}
	command = validCreateCommand()
	command.Token.DeliveryArtifactRef = "iam-delivery://tenant/not-a-uuid/artifact/secret"
	if err := command.Validate(); err == nil {
		t.Fatal("noncanonical delivery artifact reference was accepted")
	}
}
