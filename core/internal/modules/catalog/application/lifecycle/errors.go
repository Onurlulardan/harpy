package lifecycle

import "errors"

var (
	ErrInvalid           = errors.New("catalog lifecycle request is invalid")
	ErrForbidden         = errors.New("catalog lifecycle request is forbidden")
	ErrNotFound          = errors.New("catalog lifecycle resource was not found")
	ErrConflict          = errors.New("catalog lifecycle state conflicts")
	ErrOwnerNotV2        = errors.New("CATALOG_OWNER_NOT_V2")
	ErrOwnerReceiptStale = errors.New("CATALOG_OWNER_RECEIPT_STALE")
	ErrDependencyBlocked = errors.New("CATALOG_DEPENDENCY_BLOCKED")
	ErrBackupBlocked     = errors.New("CATALOG_BACKUP_BLOCKED")
	ErrOutcomeUnknown    = errors.New("CATALOG_OUTCOME_UNKNOWN")
	ErrResumeInterrupted = errors.New("CATALOG_RESUME_INTERRUPTED")
	ErrDemoPlanStale     = errors.New("CATALOG_DEMO_PLAN_STALE")
	ErrUnavailable       = errors.New("catalog lifecycle authority is unavailable")
)
