package transaction

import "context"

func callCredentialScope(
	ctx context.Context,
	scope CredentialTenantScope,
	execute func(context.Context, CredentialTenantScope) error,
) (callbackErr error, panicValue any) {
	defer func() { panicValue = recover() }()
	callbackErr = execute(markCredentialScopeContext(ctx, scope), scope)
	return callbackErr, nil
}
