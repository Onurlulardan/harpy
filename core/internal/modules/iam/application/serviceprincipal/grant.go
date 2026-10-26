package serviceprincipal

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type GrantCommand struct {
	GrantID          uuid.UUID
	PrincipalID      uuid.UUID
	ExpectedRevision int64
	Capability       CapabilityKey
	Scope            ResourceScope
	ValidFrom        time.Time
	ValidUntil       time.Time
	ManifestSHA256   [32]byte
	Evidence         MutationEvidence
}

func (command GrantCommand) Validate() error {
	if command.GrantID == uuid.Nil || command.PrincipalID == uuid.Nil ||
		command.ExpectedRevision < 0 || command.ValidFrom.IsZero() ||
		(!command.ValidUntil.IsZero() && !command.ValidUntil.After(command.ValidFrom)) ||
		command.ManifestSHA256 == [32]byte{} || command.Scope.Validate() != nil ||
		command.Evidence.Validate() != nil {
		return ErrInvalidGrant
	}
	_, err := ParseCapabilityKey(string(command.Capability))
	return err
}

func (commands *Commands) Grant(
	ctx context.Context, scope transaction.TenantScope, command GrantCommand,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil ||
		!commands.catalog.Allows(command.Capability, command.Scope) ||
		!mutationScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidGrant
	}
	result, err := commands.port.Grant(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validMutationResult(result, command.GrantID) ||
		result.Revision != command.ExpectedRevision+1 {
		return MutationResult{}, ErrEvidenceMissing
	}
	return commands.finish(
		ctx, scope, command.Evidence, result,
		"iam.capability_grant.write", "iam.capability_grant.written",
	)
}
