package profile

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func commandScopeUsable(
	ctx context.Context,
	scope transaction.TenantScope,
	evidence MutationEvidence,
) bool {
	return ctx != nil && ctx.Err() == nil && evidence.Validate() == nil &&
		!transaction.NetworkAllowed(ctx) && transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.Command) == nil &&
		scope.ActorID() == evidence.ActorID
}

func stableID(id uuid.UUID) bool {
	return id != uuid.Nil && id.Variant() == uuid.RFC4122
}

func validResult(result MutationResult, id uuid.UUID) bool {
	return result.Validate() == nil && result.ID == id
}
