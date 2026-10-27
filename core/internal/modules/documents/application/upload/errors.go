package upload

import "errors"

var (
	ErrInvalid         = errors.New("documents upload input is invalid")
	ErrForbidden       = errors.New("documents upload operation is forbidden")
	ErrNotFound        = errors.New("documents upload was not found")
	ErrConflict        = errors.New("documents current version conflicts")
	ErrInProgress      = errors.New("documents upload is in progress")
	ErrScanUnavailable = errors.New("documents scanner is unavailable")
	ErrContentRejected = errors.New("documents content was rejected")
	ErrQuota           = errors.New("documents storage limit is exceeded")
	ErrUnavailable     = errors.New("documents upload dependency is unavailable")
)

type EffectFailureKind string

const (
	FailureTransient EffectFailureKind = "transient"
	FailureRejected  EffectFailureKind = "rejected"
	FailureInvalid   EffectFailureKind = "invalid"
)
