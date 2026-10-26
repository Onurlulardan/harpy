package api

import "context"

type RecoveryIdentity struct {
	EnvironmentID   string
	TenantID        string
	TenantSetSHA256 [32]byte
}

type RecoveryInventoryItem struct {
	VersionID        string
	ObjectKey        string
	ObjectVersion    string
	ObjectGeneration int64
	CiphertextSHA256 [32]byte
	CiphertextBytes  int64
	KeyID            string
}

type RecoveryLocatorRemap struct {
	RestoreOperationID     string
	RecoverySetID          string
	Item                   RecoveryInventoryItem
	TargetObjectKey        string
	TargetObjectVersion    string
	TargetObjectGeneration int64
}

type RecoveryTransientDisposition struct {
	CancelledUploads  int64
	ExpiredArchives   int64
	CancelledArchives int64
}

type Recovery interface {
	BeginHold(context.Context, string) error
	ReadInventoryAtSnapshot(
		context.Context, string, string,
	) ([]RecoveryInventoryItem, [32]byte, error)
	ListInventory(
		context.Context, string,
	) ([]RecoveryInventoryItem, [32]byte, error)
	RecordCopy(
		context.Context, string, RecoveryInventoryItem, string,
	) error
	CompleteHold(context.Context, string, [32]byte, int64) error
	ReleaseHold(context.Context, string, [32]byte) error
	FailHold(context.Context, string, string) error
	ApplyLocator(context.Context, RecoveryLocatorRemap) error
	TerminalizeTransients(
		context.Context, string,
	) (RecoveryTransientDisposition, error)
	ValidateInventory(
		[]RecoveryInventoryItem, [32]byte,
	) ([]RecoveryInventoryItem, error)
}
