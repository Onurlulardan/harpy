package transaction

import "context"

func callProviderScope(
	ctx context.Context,
	scope ProviderTenantScope,
	execute func(context.Context, ProviderTenantScope) error,
) (callbackErr error, panicValue any) {
	defer func() { panicValue = recover() }()
	callbackErr = execute(markProviderScopeContext(ctx, scope), scope)
	return callbackErr, nil
}
