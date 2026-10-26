package tenantsettings

import "errors"

var (
	ErrAuthentication = errors.New("tenant settings authentication failed")
	ErrForbidden      = errors.New("tenant settings operation is forbidden")
	ErrInvalid        = errors.New("tenant settings input is invalid")
	ErrPrecondition   = errors.New("tenant settings revision precondition is required")
	ErrNotFound       = errors.New("tenant settings were not found")
	ErrConflict       = errors.New("tenant settings state conflicts")
	ErrUnavailable    = errors.New("tenant settings are unavailable")
)
