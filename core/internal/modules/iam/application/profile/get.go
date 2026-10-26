package profile

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) GetSelf(
	ctx context.Context,
	scope transaction.TenantScope,
	evidence SelfEvidence,
) (SelfRecord, error) {
	if commands == nil || evidence.Validate() != nil || ctx == nil || ctx.Err() != nil ||
		transaction.NetworkAllowed(ctx) || !transaction.ContextOwnsScope(ctx, scope) ||
		!selfReadScopeActive(scope) ||
		scope.ActorID() != evidence.ActorID {
		return SelfRecord{}, ErrInvalidCommand
	}
	record, err := commands.port.GetSelf(ctx, scope, evidence)
	if err != nil {
		return SelfRecord{}, err
	}
	if record.Validate(evidence.ActorID) != nil {
		return SelfRecord{}, ErrEvidenceMissing
	}
	return record, nil
}

func selfReadScopeActive(scope transaction.TenantScope) bool {
	return transaction.RequireActive(scope, transaction.SnapshotQuery) == nil ||
		transaction.RequireActive(scope, transaction.Command) == nil
}
