package recovery

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalid     = errors.New("documents recovery input is invalid")
	ErrConflict    = errors.New("documents recovery state conflicts")
	ErrUnavailable = errors.New("documents recovery is unavailable")
)

type Identity struct {
	EnvironmentID   uuid.UUID
	TenantID        uuid.UUID
	TenantSetSHA256 [32]byte
}

type InventoryItem struct {
	VersionID        uuid.UUID
	ObjectKey        string
	ObjectVersion    string
	ObjectGeneration int64
	CiphertextSHA256 [32]byte
	CiphertextBytes  int64
	KeyID            string
}

type CopyReceipt struct {
	RecoverySetID       uuid.UUID
	Item                InventoryItem
	TargetObjectKey     string
	TargetObjectVersion string
}

type LocatorRemap struct {
	RestoreOperationID     uuid.UUID
	RecoverySetID          uuid.UUID
	Item                   InventoryItem
	TargetObjectKey        string
	TargetObjectVersion    string
	TargetObjectGeneration int64
}

type TransientDisposition struct {
	CancelledUploads  int64
	ExpiredArchives   int64
	CancelledArchives int64
}

type Hold struct {
	State     string
	StartedAt time.Time
}
