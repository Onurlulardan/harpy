package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestPublishRequiresBoundActorTenantScope(t *testing.T) {
	command := PublishCommand{
		Change: SchemaChange{
			Schema:                 fixtureCommandSchema(t, FieldActive),
			ExpectedActiveSchemaID: fixtureActiveSchema,
		},
		Evidence: fixtureEvidence(),
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid publish rejected: %v", err)
	}
	_, err := fixtureCommands(t).Publish(
		context.Background(), transaction.TenantScope{}, command,
	)
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("unbound publish error = %v", err)
	}
}
