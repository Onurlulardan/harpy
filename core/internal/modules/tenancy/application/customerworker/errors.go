package customerworker

import "errors"

var (
	ErrInvalid     = errors.New("customer provisioning worker input is invalid")
	ErrNoWork      = errors.New("customer provisioning worker has no work")
	ErrUnavailable = errors.New("customer provisioning worker is unavailable")
)
