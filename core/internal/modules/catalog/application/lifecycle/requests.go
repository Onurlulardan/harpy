package lifecycle

import "github.com/google/uuid"

type InstallRequest struct {
	PackageID      uuid.UUID
	Version        string
	IdempotencyKey string
}

type UpgradeRequest struct {
	PackageID      uuid.UUID
	Version        string
	IdempotencyKey string
}

type RetryRequest struct {
	OperationID    uuid.UUID
	IdempotencyKey string
}

type CancelRequest struct {
	OperationID    uuid.UUID
	IdempotencyKey string
}

type PackageCommand struct {
	PackageID      uuid.UUID
	IdempotencyKey string
}

type DemoInstallRequest struct {
	InstallationID uuid.UUID
	IdempotencyKey string
}

type DemoRemoveRequest struct {
	InstallationID     uuid.UUID
	ExpectedPlanSHA256 string
	IdempotencyKey     string
}

type PurgeRequest struct {
	OperationID       uuid.UUID
	PackageID         uuid.UUID
	BackupReceiptID   uuid.UUID
	ApprovalReceiptID uuid.UUID
	RequestedBy       uuid.UUID
	StepUpAtUnix      int64
	Reason            string
	ConfirmationSlug  string
	IdempotencyKey    string
}

type PurgeApprovalRequest struct {
	InstallationID         uuid.UUID
	OperationID            uuid.UUID
	BackupReceiptID        uuid.UUID
	RequestedBy            uuid.UUID
	StepUpAtUnix           int64
	Reason                 string
	ConfirmationSlug       string
	PurgeIdempotencyKey    string
	ApprovalIdempotencyKey string
}
