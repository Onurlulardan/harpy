package transaction

import "errors"

var (
	ErrInvalidScopeRequest           = errors.New("tenant scope request is invalid")
	ErrInvalidGlobalScopeRequest     = errors.New("global scope request is invalid")
	ErrInvalidCredentialScopeRequest = errors.New("credential tenant scope request is invalid")
	ErrInvalidProviderScopeRequest   = errors.New("provider tenant scope request is invalid")
	ErrNestedScope                   = errors.New("nested runtime scope is forbidden")
	ErrScopeClosed                   = errors.New("tenant scope is closed")
	ErrScopeMode                     = errors.New("tenant scope mode is incompatible")
	ErrScopeOwner                    = errors.New("tenant scope owner is unauthorized")
	ErrCredentialPurpose             = errors.New("credential scope purpose is unauthorized")
	ErrCredentialResolver            = errors.New("credential target resolver is unavailable")
	ErrCredentialTarget              = errors.New("credential target is unavailable")
	ErrProviderResolver              = errors.New("provider target resolver is unavailable")
	ErrProviderTarget                = errors.New("provider target is unavailable")
	ErrTenantMissing                 = errors.New("tenant is missing")
	ErrTenantInactive                = errors.New("tenant is inactive")
	ErrTenantIdentity                = errors.New("tenant schema identity is invalid")
	ErrWriterAuthority               = errors.New("tenant writer authority is unavailable")
	ErrGlobalIdentity                = errors.New("global environment identity is unavailable")
	ErrGlobalAuthority               = errors.New("global billing authority is unavailable")
	ErrScopeDatabase                 = errors.New("tenant scope database operation failed")
	ErrGlobalScopeDatabase           = errors.New("global scope database operation failed")
	ErrRollbackFailed                = errors.New("runtime transaction rollback failed")
	ErrCommitRejected                = errors.New("runtime transaction was not committed")
)

type AmbiguousCommitError struct{}

func (AmbiguousCommitError) Error() string {
	return "runtime transaction commit outcome is ambiguous"
}
