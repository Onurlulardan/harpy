package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestBackfillRequiresBoundActorTenantScope(t *testing.T) {
	schema := fixtureCommandSchema(t, FieldActive)
	document, err := DecodeDocument(schema, []byte(`{}`), nil, nil)
	if err != nil {
		t.Fatalf("create target document: %v", err)
	}
	command := BackfillCommand{
		ResourceID: fixtureResourceID, ExpectedSourceSchemaID: fixtureActiveSchema,
		ExpectedRevision: 4, Target: document, Evidence: fixtureEvidence(),
	}
	if err = command.Validate(); err != nil {
		t.Fatalf("valid backfill rejected: %v", err)
	}
	_, err = fixtureCommands(t).Backfill(
		context.Background(), transaction.TenantScope{}, command,
	)
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("unbound backfill error = %v", err)
	}
}
