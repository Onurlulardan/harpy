package invoice

import "errors"

var (
	ErrConfiguration   = errors.New("billing invoice finalizer is misconfigured")
	ErrRequest         = errors.New("billing invoice finalization request is invalid")
	ErrSnapshot        = errors.New("billing invoice draft snapshot is invalid")
	ErrState           = errors.New("billing invoice state conflicts")
	ErrOperationReused = errors.New("billing invoice operation was reused")
)
