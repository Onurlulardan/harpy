package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CredentialScope = platformtx.CredentialTenantScope

func validCredentialScope(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	purpose platformtx.CredentialPurpose,
	locator platformtx.CredentialLocatorKind,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsCredentialScope(ctx, scope) &&
		platformtx.RequireCredentialPurpose(scope, purpose) == nil &&
		scope.LocatorKind() == locator && scope.TenantID() != [16]byte{} &&
		scope.SourceRevision() > 0 && scope.LocatorBindingSHA256() != [32]byte{}
}
