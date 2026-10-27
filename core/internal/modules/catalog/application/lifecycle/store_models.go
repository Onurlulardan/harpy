package lifecycle

import (
	"time"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

type PurgeAuthority = catalogcomponentv2.PurgeAuthority

type PurgeApprovalReceipt struct {
	ID, TenantID, InstallationID, PackageID, OperationID  string
	RequestedBy, ApprovedBy                               string
	BackupReceiptID, BackupSHA256                         string
	ConfirmationSlug, Reason, PurgeIdempotencyKey         string
	RequestSHA256, ApprovalIdempotencyKey, ApprovalSHA256 string
	StepUpAuthenticatedAt, ApprovedAt, ExpiresAt          time.Time
	State                                                 string
	ConsumedOperationID, ConsumedAuthoritySHA256          string
	ConsumedAt                                            time.Time
}

type PurgeApprovalMutation struct {
	Receipt PurgeApprovalReceipt
}

type StartMutation struct {
	Installation       lifecycle.Installation
	Operation          lifecycle.Operation
	PreviousOperation  *uuid.UUID
	PreviousState      *lifecycle.InstallationState
	ActorID            uuid.UUID
	IdempotencyKey     string
	RequestSHA256      string
	PurgeAuthority     *PurgeAuthority
	OccurredAt         time.Time
	MutateInstallation bool
}

type InstallationEvent struct {
	Type, Version, IdempotencyKey string
	PlanSHA256                    string
	Installation                  lifecycle.Installation
	OperationID                   packagecontract.StableID
	OccurredAt                    time.Time
}

type OperationHandle struct {
	Operation lifecycle.Operation
	Replayed  bool
}

type OperationMutation struct {
	Operation        lifecycle.Operation
	ExpectedRevision uint64
	FailureCode      string
	OwnerReceipts    []lifecycle.OwnerReceipt
}

type InstallationMutation struct {
	Installation     lifecycle.Installation
	OperationID      packagecontract.StableID
	ExpectedRevision uint64
	ActorID          uuid.UUID
	IdempotencyKey   string
	RequestSHA256    string
	OccurredAt       time.Time
	OwnerReceipts    []lifecycle.OwnerReceipt
}

type DependencyBlocker struct {
	PackageID packagecontract.PackageID
	Version   packagecontract.Version
	Required  bool
}

type PageRequest struct {
	Limit int
	After uuid.UUID
}

type InstallationPage struct {
	Items []lifecycle.Installation
	Next  uuid.UUID
}
