package tenantmigrationexecutor

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/google/uuid"
)

const (
	ReadyTTL         = 30 * time.Second
	ReadyHeartbeat   = 10 * time.Second
	LeaseTTL         = 60 * time.Second
	LeaseHeartbeat   = 20 * time.Second
	IdlePollInterval = 500 * time.Millisecond
	MaximumAttempts  = 8
)

var (
	ErrAmbiguousOutcome = errors.New("tenant migration outcome is ambiguous")
	ErrCleanupRequired  = errors.New("tenant migration schema requires offline cleanup")
)

type Release struct {
	ID                      uuid.UUID
	DeploymentMode          string
	RuntimeGeneration       int64
	MigrationManifestSHA256 string
}

type ReadyReceipt struct {
	ProfileInstanceID       uuid.UUID
	ReleaseID               uuid.UUID
	EnvironmentID           uuid.UUID
	DeploymentMode          string
	RuntimeGeneration       int64
	DatabaseGeneration      int64
	MigrationManifestSHA256 string
	ReadyUntil              time.Time
}

type Claim struct {
	JobID, OperationID, TenantID, MigrationRunID uuid.UUID
	PreviousRunID, LeaseID, ProfileInstanceID    uuid.UUID
	SchemaName, StepPlanSHA256, Origin           string
	ReleaseID, EnvironmentID                     uuid.UUID
	DeploymentMode, MigrationManifestSHA256      string
	Attempt                                      int
	RuntimeGeneration, DatabaseGeneration        int64
	LeaseUntil                                   time.Time
}

type MigrationReceipt struct {
	ReceiptSHA256, EvidenceSHA256 string
}

type Failure struct {
	Outcome, Code  string
	ReceiptID      uuid.UUID
	ReceiptSHA256  string
	EvidenceSHA256 string
}

type Store interface {
	RegisterReady(context.Context, uuid.UUID, Release, time.Duration) (ReadyReceipt, error)
	HeartbeatReady(context.Context, uuid.UUID, time.Duration) (time.Time, error)
	RetireReady(context.Context, uuid.UUID) error
	Claim(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Duration) (Claim, bool, error)
	Heartbeat(context.Context, Claim, time.Duration) (time.Time, error)
	Complete(context.Context, Claim, uuid.UUID, MigrationReceipt) error
	Fail(context.Context, Claim, Failure) error
}

type Migrator interface {
	Migrate(context.Context, Claim) (MigrationReceipt, error)
}

type Admission interface {
	Admit() (admission.Permit, bool)
	Draining() bool
}
