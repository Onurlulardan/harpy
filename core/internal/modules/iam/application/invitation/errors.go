package invitation

import "errors"

var (
	ErrInvalid          = errors.New("IAM invitation input is invalid")
	ErrUnavailable      = errors.New("IAM invitation store is unavailable")
	ErrNotFound         = errors.New("IAM invitation was not found")
	ErrStale            = errors.New("IAM invitation version or state is stale")
	ErrDenied           = errors.New("IAM invitation authority is denied")
	ErrConflict         = errors.New("IAM invitation conflicts with current state")
	ErrRecoveryRequired = errors.New("IAM profile recovery is required")
	ErrEffects          = errors.New("IAM invitation transactional effects failed")
)
