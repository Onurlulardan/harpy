package serviceprincipal

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type CreateCommand struct {
	ID          uuid.UUID
	Key         string
	DisplayName string
	Evidence    MutationEvidence
}

func (command CreateCommand) Validate() error {
	if command.ID == uuid.Nil || len(command.Key) < 3 || len(command.Key) > 100 ||
		!capabilityKeyPattern.MatchString(command.Key) ||
		len(command.DisplayName) < 1 || len(command.DisplayName) > 100 ||
		strings.TrimSpace(command.DisplayName) != command.DisplayName ||
		command.Evidence.Validate() != nil {
		return ErrInvalidPrincipal
	}
	return nil
}

func (commands *Commands) Create(
	ctx context.Context,
	scope transaction.TenantScope,
	command CreateCommand,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil ||
		!mutationScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidPrincipal
	}
	result, err := commands.port.Create(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validMutationResult(result, command.ID) {
		return MutationResult{}, ErrEvidenceMissing
	}
	return commands.finish(
		ctx, scope, command.Evidence, result,
		"iam.service_principal.create", "iam.service_principal.created",
	)
}
