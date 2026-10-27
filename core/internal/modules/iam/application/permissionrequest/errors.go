package permissionrequest

import "errors"

var (
	ErrInvalid         = errors.New("IAM permission request input is invalid")
	ErrUnavailable     = errors.New("IAM permission request store is unavailable")
	ErrNotFound        = errors.New("IAM permission request was not found")
	ErrStale           = errors.New("IAM permission request version or state is stale")
	ErrDenied          = errors.New("IAM permission request authority is denied")
	ErrConflict        = errors.New("IAM permission request operation conflicts")
	ErrDecisionEffects = errors.New("IAM permission request decision effects failed")
)
