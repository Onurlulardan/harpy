package identity

import "errors"

var (
	ErrInvalidCommand     = errors.New("IAM identity command is invalid")
	ErrInvalidQuery       = errors.New("IAM identity query is invalid")
	ErrInvalidRecord      = errors.New("IAM identity record is invalid")
	ErrStorage            = errors.New("IAM identity storage failed")
	ErrChangeSink         = errors.New("IAM identity audit/outbox sink failed")
	ErrBulkDeleteConflict = errors.New("IAM identity bulk delete replay conflicts")
)
