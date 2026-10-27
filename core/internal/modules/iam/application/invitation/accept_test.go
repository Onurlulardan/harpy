package invitation

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func TestAcceptCommandRequiresHashedCredentialAndIdempotencyIdentity(t *testing.T) {
	profile := []byte("{}")
	command := AcceptCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1}, ID: uuid.New(),
		ExpectedRevision: 1, Token: validTokenProof(),
		AcceptedUserID: uuid.New(), AcceptedIdentifierID: uuid.New(),
		AccessModeID: uuid.New(), RoleAssignmentID: uuid.New(),
		Profile: RegistrationProfile{
			SchemaID: uuid.New(), Values: profile, SHA256: sha256.Sum256(profile),
		},
		Password: RegistrationPassword{Salt: [16]byte{1}, Tag: [32]byte{2}},
		Directory: DirectoryProjection{
			LookupKeyID: fixtureTokenKeyID(), HMAC: [32]byte{3},
			SourceSHA256: [32]byte{4},
		},
		ExpectedAuthorizationVersion: 1,
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid accept command: %v", err)
	}
	command.Token.SecretHMAC = [32]byte{}
	if err := command.Validate(); err == nil {
		t.Fatal("missing invitation HMAC proof was accepted")
	}
	command.Token = validTokenProof()
	command.OperationID = uuid.Nil
	if err := command.Validate(); err == nil {
		t.Fatal("missing idempotency identity was accepted")
	}
}
