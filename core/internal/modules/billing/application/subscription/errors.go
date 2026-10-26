package subscriptionapp

import "errors"

var (
	ErrConfiguration       = errors.New("billing subscription application is not configured")
	ErrRequest             = errors.New("billing subscription request is invalid")
	ErrSnapshot            = errors.New("billing subscription snapshot is invalid")
	ErrStateConflict       = errors.New("billing subscription state conflicts with operation")
	ErrStaleRevision       = errors.New("billing subscription revision is stale")
	ErrIdempotencyReused   = errors.New("billing subscription operation identity was reused")
	ErrCancellationMissing = errors.New("billing subscription cancellation schedule is missing")
	ErrCancellationExists  = errors.New("billing subscription cancellation is already scheduled")
	ErrChangeExists        = errors.New("billing subscription change is already scheduled")
	ErrTrialClaimed        = errors.New("billing subscription trial policy was already claimed")
	ErrProviderFactReused  = errors.New("billing provider receipt or fact was reused")
	ErrBoundaryNotDue      = errors.New("billing subscription boundary is not due")
	ErrBoundaryMismatch    = errors.New("billing subscription boundary does not match durable state")
)
