package postrestore

import "errors"

var (
	ErrInvalid     = errors.New("IAM post-restore input is invalid")
	ErrDenied      = errors.New("IAM post-restore operation is denied")
	ErrUnavailable = errors.New("IAM post-restore runtime is unavailable")
)
