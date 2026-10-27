package lifecycle

import "errors"

var (
	ErrInvalid          = errors.New("catalog lifecycle input is invalid")
	ErrTransition       = errors.New("catalog lifecycle transition is forbidden")
	ErrImmutable        = errors.New("catalog lifecycle immutable value changed")
	ErrOwnerNotReady    = errors.New("catalog component owner is not ready")
	ErrReceiptMismatch  = errors.New("catalog component receipt does not match the plan")
	ErrCancellationLate = errors.New("catalog operation can no longer be cancelled")
)
