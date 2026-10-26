package transaction

import "context"

func callScope(
	ctx context.Context,
	scope TenantScope,
	execute func(context.Context, TenantScope) error,
) (callbackErr error, panicValue any) {
	defer func() {
		panicValue = recover()
	}()
	callbackErr = execute(markScopeContext(ctx, scope), scope)
	return callbackErr, nil
}
