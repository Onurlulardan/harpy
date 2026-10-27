package transaction

import "github.com/Onurlulardan/harpy/core/internal/platform/postgres"

const asyncAuthorityBindingOwner = "platform_async"

func BindAuthority[T any](
	scope TenantScope,
	authority OwnerAuthority,
	binder postgres.AuthorityBinder[T],
) (T, error) {
	var zero T
	if !authority.OwnsScope(scope) || scope.Mode() != Command ||
		binder.Owner() != asyncAuthorityBindingOwner {
		return zero, postgres.ErrAuthorityBindingUnauthorized
	}
	return postgres.BindAuthority(scope.state.transaction, binder)
}

func BindProviderAuthority[T any](
	scope ProviderTenantScope,
	authority OwnerAuthority,
	binder postgres.AuthorityBinder[T],
) (T, error) {
	var zero T
	if !authority.OwnsProviderScope(scope) ||
		binder.Owner() != asyncAuthorityBindingOwner {
		return zero, postgres.ErrAuthorityBindingUnauthorized
	}
	return postgres.BindAuthority(scope.state.transaction, binder)
}
