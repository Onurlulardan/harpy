package systemtransfer

import "context"

func importWorkerScopeUsable(ctx context.Context, scope tenantScope) bool {
	return commandScopeUsable(ctx, scope, "iam.record_import.execute")
}

func exportWorkerScopeUsable(ctx context.Context, scope tenantScope) bool {
	return commandScopeUsable(ctx, scope, exportWorkerOperation)
}

func commandScopeUsable(ctx context.Context, scope tenantScope, operation string) bool {
	return ownedScopeUsable(ctx, scope, commandScopeMode) &&
		scope.OperationID() == operation
}
