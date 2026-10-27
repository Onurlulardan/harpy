package postgres

import "errors"

var (
	ErrTenantMissing                = errors.New("tenant registry row is missing")
	ErrTenantInactive               = errors.New("tenant registry row is inactive")
	ErrTenantSchemaInvalid          = errors.New("tenant schema name is invalid")
	ErrTenantIdentityMismatch       = errors.New("tenant schema identity does not match registry")
	ErrGlobalIdentityMismatch       = errors.New("global environment identity does not match registry")
	ErrAuthorityMismatch            = errors.New("runtime writer authority does not match operation")
	ErrOwnerBindingInvalid          = errors.New("postgres owner binding is invalid")
	ErrOwnerBindingUnauthorized     = errors.New("postgres owner binding is not authorized")
	ErrAuthorityBindingInvalid      = errors.New("postgres authority binding is invalid")
	ErrAuthorityBindingUnauthorized = errors.New("postgres authority binding is not authorized")
)
