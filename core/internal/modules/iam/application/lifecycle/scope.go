package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func tenantCommandUsable(
	ctx context.Context, scope platformtx.TenantScope, evidence MutationEvidence,
) bool {
	return ctx != nil && ctx.Err() == nil && evidence.Validate() == nil &&
		!platformtx.NetworkAllowed(ctx) && platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil &&
		scope.ActorID() == evidence.ActorID
}

func tenantSnapshotUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
}

func credentialUsable(
	ctx context.Context, scope platformtx.CredentialTenantScope,
	evidence CredentialEvidence, purpose platformtx.CredentialPurpose,
) bool {
	return ctx != nil && ctx.Err() == nil && evidence.Validate() == nil &&
		!platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsCredentialScope(ctx, scope) &&
		platformtx.RequireCredentialPurpose(scope, purpose) == nil
}

func credentialLookupUsable(
	ctx context.Context, scope platformtx.CredentialTenantScope,
	purpose platformtx.CredentialPurpose,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsCredentialScope(ctx, scope) &&
		platformtx.RequireCredentialPurpose(scope, purpose) == nil
}
