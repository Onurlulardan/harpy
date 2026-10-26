package customerprovisioning

import "errors"

var (
	ErrInvalid        = errors.New("customer provisioning request is invalid")
	ErrAuthentication = errors.New("customer provisioning authentication failed")
	ErrForbidden      = errors.New("customer provisioning is forbidden")
	ErrModeDenied     = errors.New("customer provisioning deployment mode is denied")
	ErrConflict       = errors.New("customer provisioning operation conflicts")
	ErrNotFound       = errors.New("customer provisioning operation was not found")
	ErrUnavailable    = errors.New("customer provisioning is unavailable")
)
