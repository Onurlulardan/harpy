package lifecycle

import "errors"

var (
	ErrInvalid     = errors.New("IAM lifecycle input is invalid")
	ErrDenied      = errors.New("IAM lifecycle operation is denied")
	ErrNotFound    = errors.New("IAM lifecycle resource was not found")
	ErrConflict    = errors.New("IAM lifecycle operation conflicts")
	ErrStale       = errors.New("IAM lifecycle revision changed")
	ErrUnavailable = errors.New("IAM lifecycle runtime is unavailable")
	ErrEffects     = errors.New("IAM lifecycle effects failed")
)
