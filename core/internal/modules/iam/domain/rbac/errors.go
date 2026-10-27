package rbac

import "errors"

var (
	ErrInvalidMutation = errors.New("IAM RBAC mutation is invalid")
	ErrInvalidQuery    = errors.New("IAM RBAC query is invalid")
	ErrNotFound        = errors.New("IAM RBAC resource was not found")
	ErrConflict        = errors.New("IAM RBAC operation conflicts with current state")
	ErrDenied          = errors.New("IAM RBAC authority is denied")
	ErrStorage         = errors.New("IAM RBAC storage failed")
	ErrSink            = errors.New("IAM RBAC audit/outbox sink failed")
)
