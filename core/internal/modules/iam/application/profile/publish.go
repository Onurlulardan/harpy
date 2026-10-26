package profile

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PublishCommand struct {
	Change   SchemaChange
	Evidence MutationEvidence
}

func (command PublishCommand) Validate() error {
	if command.Change.Validate() != nil || command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (commands *Commands) Publish(
	ctx context.Context,
	scope transaction.TenantScope,
	command PublishCommand,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil ||
		!commandScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidCommand
	}
	result, err := commands.port.Publish(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validResult(result, command.Change.Schema.ID()) {
		return MutationResult{}, ErrEvidenceMissing
	}
	return result, nil
}
