package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRetireRequiresBoundActorTenantScope(t *testing.T) {
	target := RetirementTarget{
		Resource: ResourceUser, FieldID: fixtureFieldID,
		ExpectedFieldRevision: 1,
	}
	command := RetireCommand{
		Change: SchemaChange{
			Schema:                 fixtureCommandSchema(t, FieldRetired),
			ExpectedActiveSchemaID: fixtureActiveSchema,
		},
		Target: target, Evidence: fixtureEvidence(),
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid retirement rejected: %v", err)
	}
	_, err := fixtureCommands(t).Retire(
		context.Background(), transaction.TenantScope{}, command,
	)
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("unbound retirement error = %v", err)
	}
}

func TestReferenceEvidenceFailsClosedOnLiveReference(t *testing.T) {
	target := RetirementTarget{
		Resource: ResourceUser, FieldID: fixtureFieldID,
		ExpectedFieldRevision: 1,
	}
	evidence := ReferenceEvidence{
		Target: target, ManifestSHA256: [32]byte{1},
		CheckedOwners: 1, LiveReferences: 1,
	}
	if !errors.Is(evidence.Validate(target), ErrReferencePresent) {
		t.Fatal("live reference was accepted")
	}
}
