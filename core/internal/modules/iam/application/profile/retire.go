package profile

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type RetireCommand struct {
	Change     SchemaChange
	Target     RetirementTarget
	References ReferenceEvidence
	Evidence   MutationEvidence
}

func (command RetireCommand) Validate() error {
	if command.Change.Validate() != nil || command.Target.Validate() != nil ||
		command.Evidence.Validate() != nil ||
		command.Target.Resource != command.Change.Schema.Resource() {
		return ErrInvalidCommand
	}
	field, found := command.Change.Schema.Field(command.Target.FieldID)
	if !found || field.State() != FieldRetired {
		return ErrInvalidCommand
	}
	if command.References != (ReferenceEvidence{}) &&
		command.References.Validate(command.Target) != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (commands *Commands) Retire(
	ctx context.Context,
	scope transaction.TenantScope,
	command RetireCommand,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil ||
		!commandScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidCommand
	}
	references, err := commands.references.AssertRetirable(ctx, scope, command.Target)
	if err != nil {
		return MutationResult{}, err
	}
	if err = references.Validate(command.Target); err != nil {
		return MutationResult{}, err
	}
	command.References = references
	result, err := commands.port.Retire(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validResult(result, command.Change.Schema.ID()) {
		return MutationResult{}, ErrEvidenceMissing
	}
	return result, nil
}
