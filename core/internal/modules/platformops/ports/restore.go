package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

type RestoreRequestRecord struct {
	OperationID, RecoverySetID, EnvironmentID string
	RequestedByTenantID, RequestedByActorID   string
	RequestedBySessionID                      string
	Disposition                               recoveryset.TargetDisposition
	PlanSHA256                                string
	PlanBytes                                 []byte
	Safety                                    recoveryset.SafetyEvidence
	CreatedAt                                 time.Time
}

type RestoreRecord struct {
	OperationID         string
	RecoverySetID       string
	EnvironmentID       string
	Disposition         recoveryset.TargetDisposition
	State               string
	PlanSHA256          string
	ReceiptSHA256       string
	NewDatabaseSystemID string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CompletedAt         *time.Time
}

type RestoreStore interface {
	RequestRestore(context.Context, RestoreRequestRecord) error
	ResolveRestore(context.Context, string) (RestoreRecord, error)
	ClaimRestorePlan(context.Context, RecoveryArtifactClaim) (RecoveryArtifactRecord, error)
}
