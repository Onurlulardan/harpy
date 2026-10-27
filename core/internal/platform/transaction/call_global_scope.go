package transaction

import "context"

func callGlobalScope(
	ctx context.Context,
	scope GlobalScope,
	execute func(context.Context, GlobalScope) error,
) (callbackErr error, panicValue any) {
	defer func() {
		panicValue = recover()
	}()
	callbackErr = execute(markGlobalScopeContext(ctx, scope), scope)
	return callbackErr, nil
}
