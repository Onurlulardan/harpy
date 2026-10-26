package tenantmigration

import "errors"

var (
	ErrUnavailable      = errors.New("tenant migration application is unavailable")
	ErrExecutorNotReady = errors.New("tenant migration executor is not claim ready")
	ErrInvalidRequest   = errors.New("tenant migration request is invalid")
)
