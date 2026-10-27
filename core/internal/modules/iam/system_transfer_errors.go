package iam

import "errors"

var (
	ErrSystemTransferInvalid        = errors.New("IAM transfer input is invalid")
	ErrSystemTransferAuthentication = errors.New("IAM transfer authentication failed")
	ErrSystemTransferForbidden      = errors.New("IAM transfer authorization failed")
	ErrSystemTransferNotFound       = errors.New("IAM transfer was not found")
	ErrSystemTransferConflict       = errors.New("IAM transfer conflicts")
	ErrSystemTransferBusy           = errors.New("IAM transfer capacity is exhausted")
	ErrSystemTransferExpired        = errors.New("IAM transfer artifact is expired")
	ErrSystemTransferTooLarge       = errors.New("IAM transfer limit is exceeded")
	ErrSystemTransferUnavailable    = errors.New("IAM transfer is unavailable")
)
