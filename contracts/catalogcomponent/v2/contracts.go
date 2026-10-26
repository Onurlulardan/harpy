package catalogcomponentv2

import (
	"time"

	catalogcomponentv1 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
)

type Component = catalogcomponentv1.Component

const (
	BackupReady       = "READY"
	StateReady        = "READY"
	StateCooling      = "COOLING"
	StateRetired      = "RETIRED"
	StateRestored     = "RESTORED"
	StateWaiting      = "WAITING"
	StateBlocked      = "BLOCKED"
	StatePurgePending = "PURGE_PENDING"
	StatePurged       = "PURGED"
)

type LifecycleCommand struct {
	TenantID       string
	PackageID      string
	PackageSlug    string
	OperationID    string
	ActorID        string
	Component      Component
	Action         string
	Reason         string
	ReferenceEpoch uint64
	IdempotencyKey string
}

type PurgeAuthority struct {
	TenantID                    string
	PackageID                   string
	PackageSlug                 string
	OperationID                 string
	ActorID                     string
	RequestedBy                 string
	ApprovedBy                  string
	ApprovalReceiptID           string
	ApprovalSHA256              string
	BackupReceiptID             string
	BackupState                 string
	BackupSHA256                string
	ConfirmationSlug            string
	Reason                      string
	IdempotencyKey              string
	ReferenceEpoch              uint64
	AuthorizationTokenVersion   int64
	AuthorizationVersion        int64
	AuthorizationClock          int64
	AuthorizationPermissionID   string
	AuthorizationEvidenceSHA256 string
	StepUpPurposeSHA256         string
	StepUpAuthenticatedAt       time.Time
	ApprovalStepUpAt            time.Time
	ApprovedAt                  time.Time
	ApprovalExpiresAt           time.Time
	RequestedAt                 time.Time
	CoolingUntil                time.Time
	AuthoritySHA256             string
}

type PurgeCommand struct {
	Authority      PurgeAuthority
	Component      Component
	IdempotencyKey string
}

type OwnerResult struct {
	ReceiptID       string
	TenantID        string
	PackageID       string
	Component       Component
	OperationID     string
	ResourceID      string
	VersionID       string
	HeadID          string
	ReferenceEpoch  uint64
	State           string
	AuthoritySHA256 string
	ResultSHA256    string
	Ready           bool
	RetryAt         time.Time
	CompletedAt     time.Time
	Blockers        []catalogcomponentv1.Blocker
}

func (result OwnerResult) Clone() OwnerResult {
	result.Blockers = append([]catalogcomponentv1.Blocker(nil), result.Blockers...)
	return result
}
