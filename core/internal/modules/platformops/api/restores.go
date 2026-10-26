package api

import (
	"context"
	"time"
)

type TargetDisposition string

const (
	TargetExistingReadable TargetDisposition = "existing_readable"
	TargetEmpty            TargetDisposition = "empty"
	TargetUnreadable       TargetDisposition = "unreadable"
)

type RestoreOperationRequest struct {
	OperationID           string
	RecoverySetID         string
	RequestedByTenantID   string
	RequestedByActorID    string
	RequestedBySessionID  string
	PlanBytes             []byte
	PlanSigningGeneration int64
}

type PlannedRestoreRequest struct {
	OperationID, RecoverySetID                   string
	RequestedByTenantID, RequestedByActorID      string
	RequestedBySessionID                         string
	Disposition                                  TargetDisposition
	SafetyRecoverySetID, SafetyRecoverySetSHA256 string
	WaiverReason, ProbeSHA256                    string
	DestructiveConfirmed                         bool
}

type RestoreOperation struct {
	OperationID         string
	RecoverySetID       string
	EnvironmentID       string
	Disposition         TargetDisposition
	State               string
	PlanSHA256          string
	ReceiptSHA256       string
	NewDatabaseSystemID string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CompletedAt         *time.Time
}

type Restores interface {
	RequestRestore(context.Context, RestoreOperationRequest) (RestoreOperation, error)
	RequestPlannedRestore(context.Context, PlannedRestoreRequest) (RestoreOperation, error)
	ResolveRestore(context.Context, string) (RestoreOperation, error)
	ClaimRestorePlan(context.Context, RecoveryArtifactClaim) (RecoveryArtifact, error)
}
