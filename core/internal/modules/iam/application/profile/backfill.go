package profile

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type BackfillCommand struct {
	ResourceID             uuid.UUID
	ExpectedSourceSchemaID uuid.UUID
	ExpectedRevision       int64
	Target                 Document
	Evidence               MutationEvidence
}

func (command BackfillCommand) Validate() error {
	document := command.Target.CanonicalJSON()
	if !stableID(command.ResourceID) || !stableID(command.ExpectedSourceSchemaID) ||
		command.ExpectedRevision < 1 || !command.Target.Resource().Valid() ||
		!stableID(command.Target.SchemaID()) ||
		command.Target.SchemaID() == command.ExpectedSourceSchemaID ||
		len(document) < 2 || len(document) > MaximumDocumentBytes ||
		command.Target.SHA256() == [32]byte{} || command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (commands *Commands) Backfill(
	ctx context.Context,
	scope transaction.TenantScope,
	command BackfillCommand,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil ||
		!commandScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidCommand
	}
	result, err := commands.port.Backfill(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validResult(result, command.ResourceID) {
		return MutationResult{}, ErrEvidenceMissing
	}
	return result, nil
}
