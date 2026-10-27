package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestUpdateSelfRequiresActorBoundCommandScope(t *testing.T) {
	fieldInput := baseFieldInput(fieldOneID, "display_name", LogicalString)
	field := mustField(t, fieldInput)
	schema := mustSchema(t, schemaOneID, 1, field)
	document := mustDocument(
		t, schema, `{"11111111-1111-4111-8111-111111111111":"Alice"}`,
	)
	command := SelfUpdateCommand{
		OperationID: fixtureOperationID, RequestSHA256: [32]byte{1},
		ExpectedSourceSchemaID: schemaOneID, ExpectedRevision: 4,
		Target: document, ActorID: fixtureActorID,
		ExpectedTokenVersion: 2, ExpectedAuthorizationVersion: 3,
	}
	if command.Validate() != nil {
		t.Fatal("valid self update command was rejected")
	}
	_, err := fixtureCommands(t).UpdateSelf(
		context.Background(), transaction.TenantScope{}, command,
	)
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("unbound self update error = %v", err)
	}
}
