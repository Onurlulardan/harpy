package systemtransfer

import "errors"

var (
	ErrInvalidMapping          = errors.New("IAM transfer mapping is invalid")
	ErrForbiddenTarget         = errors.New("IAM transfer mapping target is forbidden")
	ErrInviteRequired          = errors.New("IAM user import requires invitation fields")
	ErrInvalidReceipt          = errors.New("IAM transfer receipt is invalid")
	ErrInvalidAttempt          = errors.New("IAM transfer attempt is invalid")
	ErrDependencies            = errors.New("IAM transfer worker dependencies are invalid")
	ErrAuthorizationCheckpoint = errors.New("IAM transfer authorization checkpoint failed")
	ErrImportCatalogPinStale   = errors.New("IAM import catalog pin is stale")
	ErrReceiptConflict         = errors.New("IAM transfer receipt conflicts with source evidence")
	ErrInvitationInvalid       = errors.New("IAM transfer invitation input is invalid")
	ErrInvitationConflict      = errors.New("IAM transfer invitation conflicts")
	ErrInvitationDenied        = errors.New("IAM transfer invitation is forbidden")
	ErrInvitationUnavailable   = errors.New("IAM transfer invitation is unavailable")
	ErrWorkerLeaseLost         = errors.New("IAM transfer worker lease is lost")
	ErrWorkerUnavailable       = errors.New("IAM transfer worker dependency is unavailable")
	ErrWorkerState             = errors.New("IAM transfer worker state is inconsistent")
)
