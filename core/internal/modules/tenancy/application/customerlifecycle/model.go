package customerlifecycle

import (
	"time"

	"github.com/google/uuid"
)

const (
	CapabilityRead       = "platform.tenants.read"
	CapabilityUpdate     = "platform.tenants.update"
	CapabilitySuspend    = "platform.tenants.suspend"
	CapabilityReactivate = "platform.tenants.reactivate"
	CapabilityOffboard   = "platform.tenants.offboard"
	CapabilityPurge      = "platform.tenants.purge"

	MutationUpdate     = "update"
	MutationSuspend    = "suspend"
	MutationReactivate = "reactivate"
	OperationOffboard  = "offboard"
	OperationPurge     = "purge"
)

type Tenant struct {
	ID               uuid.UUID
	Name             string
	Slug             string
	Status           string
	Revision         int64
	SettingsRevision int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	OffboardedAt     time.Time
	PurgedAt         time.Time
}

type Page struct {
	Items      []Tenant
	NextCursor uuid.UUID
}

type Operation struct {
	ID              uuid.UUID
	ControlTenantID uuid.UUID
	TargetTenantID  uuid.UUID
	Kind            string
	State           string
	Revision        int64
	Replay          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     time.Time
	FailureCode     string
}

type PurgeProof struct {
	RetentionReceiptID   uuid.UUID
	BackupReceiptID      uuid.UUID
	ContributorReceiptID uuid.UUID
	EvidenceSHA256       [32]byte
	VerifiedAt           time.Time
}

func (proof PurgeProof) Valid(now time.Time) bool {
	return proof.RetentionReceiptID != uuid.Nil && proof.BackupReceiptID != uuid.Nil &&
		proof.ContributorReceiptID != uuid.Nil && proof.EvidenceSHA256 != [32]byte{} &&
		!proof.VerifiedAt.IsZero() && proof.VerifiedAt.Location() == time.UTC &&
		!proof.VerifiedAt.After(now) && now.Sub(proof.VerifiedAt) <= 24*time.Hour
}

type ClaimedOperation struct {
	Operation
	SchemaName string
	Proof      PurgeProof
	WorkerID   uuid.UUID
}

type ExecutionReceipt struct {
	ID        uuid.UUID
	SHA256    [32]byte
	CreatedAt time.Time
}

func (receipt ExecutionReceipt) Valid(now time.Time) bool {
	return receipt.ID != uuid.Nil && receipt.SHA256 != [32]byte{} &&
		!receipt.CreatedAt.IsZero() && receipt.CreatedAt.Location() == time.UTC &&
		!receipt.CreatedAt.After(now)
}
