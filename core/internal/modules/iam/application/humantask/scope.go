package humantask

import (
	"context"
	"strings"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func scopeUsable(
	ctx context.Context,
	scope platformtx.TenantScope,
	request humantaskapi.VerificationRequest,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil &&
		[16]byte(scope.TenantID()) == request.TenantID &&
		strings.HasPrefix(scope.OperationID(), "workflow.human_task.")
}

func assignmentScopeUsable(
	ctx context.Context,
	scope platformtx.TenantScope,
	request humantaskapi.AssignmentRequest,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil &&
		[16]byte(scope.TenantID()) == request.TenantID &&
		strings.HasPrefix(scope.OperationID(), "workflow.human_task.")
}
