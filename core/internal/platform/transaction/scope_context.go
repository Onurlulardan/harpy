package transaction

import "context"

type scopeContextKey struct{}
type globalScopeContextKey struct{}
type credentialScopeContextKey struct{}
type credentialResolutionContextKey struct{}
type jobScopeContextKey struct{}
type providerScopeContextKey struct{}
type providerResolutionContextKey struct{}

func markScopeContext(ctx context.Context, scope TenantScope) context.Context {
	return context.WithValue(ctx, scopeContextKey{}, scope.state)
}

func nestedScope(ctx context.Context) bool {
	return scopeStateFromContext(ctx) != nil || globalScopeStateFromContext(ctx) != nil ||
		credentialScopeStateFromContext(ctx) != nil ||
		jobScopeStateFromContext(ctx) != nil || providerScopeStateFromContext(ctx) != nil ||
		credentialResolutionFromContext(ctx) || providerResolutionFromContext(ctx)
}

func markGlobalScopeContext(ctx context.Context, scope GlobalScope) context.Context {
	return context.WithValue(ctx, globalScopeContextKey{}, scope.state)
}

func ContextOwnsGlobalScope(ctx context.Context, scope GlobalScope) bool {
	return scope.state != nil && globalScopeStateFromContext(ctx) == scope.state
}

func markJobScopeContext(ctx context.Context, scope JobTenantScope) context.Context {
	return context.WithValue(ctx, jobScopeContextKey{}, scope.state)
}

func ContextOwnsJobScope(ctx context.Context, scope JobTenantScope) bool {
	return scope.state != nil && jobScopeStateFromContext(ctx) == scope.state
}

func markProviderScopeContext(
	ctx context.Context,
	scope ProviderTenantScope,
) context.Context {
	return context.WithValue(ctx, providerScopeContextKey{}, scope.state)
}

func ContextOwnsProviderScope(ctx context.Context, scope ProviderTenantScope) bool {
	return scope.state != nil && providerScopeStateFromContext(ctx) == scope.state
}

func NetworkAllowed(ctx context.Context) bool {
	return !nestedScope(ctx)
}

// ContextOwnsScope verifies that ctx was minted for this exact opaque scope.
func ContextOwnsScope(ctx context.Context, scope TenantScope) bool {
	return scope.state != nil && scopeStateFromContext(ctx) == scope.state
}

func markCredentialScopeContext(
	ctx context.Context,
	scope CredentialTenantScope,
) context.Context {
	return context.WithValue(ctx, credentialScopeContextKey{}, scope.state)
}

func ContextOwnsCredentialScope(ctx context.Context, scope CredentialTenantScope) bool {
	return scope.state != nil && credentialScopeStateFromContext(ctx) == scope.state
}

func scopeStateFromContext(ctx context.Context) *scopeState {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(scopeContextKey{}).(*scopeState)
	return state
}

func globalScopeStateFromContext(ctx context.Context) *globalScopeState {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(globalScopeContextKey{}).(*globalScopeState)
	return state
}

func credentialScopeStateFromContext(ctx context.Context) *credentialScopeState {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(credentialScopeContextKey{}).(*credentialScopeState)
	return state
}

func jobScopeStateFromContext(ctx context.Context) *jobScopeState {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(jobScopeContextKey{}).(*jobScopeState)
	return state
}

func providerScopeStateFromContext(ctx context.Context) *providerScopeState {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(providerScopeContextKey{}).(*providerScopeState)
	return state
}

func markCredentialResolutionContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, credentialResolutionContextKey{}, true)
}

func credentialResolutionFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	value, _ := ctx.Value(credentialResolutionContextKey{}).(bool)
	return value
}

func markProviderResolutionContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, providerResolutionContextKey{}, true)
}

func providerResolutionFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	value, _ := ctx.Value(providerResolutionContextKey{}).(bool)
	return value
}
