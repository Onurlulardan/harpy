package serviceprincipal

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type RevokeCommand struct {
	GrantID          uuid.UUID
	ExpectedRevision int64
	Evidence         MutationEvidence
}

func (command RevokeCommand) Validate() error {
	if command.GrantID == uuid.Nil || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidGrant
	}
	return nil
}

func (commands *Commands) Revoke(
	ctx context.Context, scope transaction.TenantScope, command RevokeCommand,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil ||
		!mutationScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidGrant
	}
	result, err := commands.port.Revoke(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validMutationResult(result, command.GrantID) ||
		result.Revision != command.ExpectedRevision+1 {
		return MutationResult{}, ErrEvidenceMissing
	}
	return commands.finish(
		ctx, scope, command.Evidence, result,
		"iam.capability_grant.revoke", "iam.capability_grant.revoked",
	)
}
