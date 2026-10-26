package policy

import (
	"context"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func requestScopeUsable(
	ctx context.Context,
	scope transaction.TenantScope,
	request restauthorization.PolicyRequest,
) bool {
	if ctx == nil || ctx.Err() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsScope(ctx, scope) ||
		transaction.RequireActive(scope, transaction.SnapshotQuery) != nil ||
		request.Validate() != nil {
		return false
	}
	tenantID, tenantErr := request.TenantID()
	actorID, actorErr := request.ActorID()
	owner, ownerErr := request.Owner()
	return tenantErr == nil && actorErr == nil && ownerErr == nil &&
		tenantID == scope.TenantID() && actorID == scope.ActorID() && owner == catalog.Owner
}
