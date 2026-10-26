package ports

import (
	"context"
	"time"
)

type RecoveryProtectionRecord struct {
	EnvironmentID   string
	State           string
	Generation      int64
	KitAcknowledged bool
	UpdatedAt       time.Time
}

type RecoveryControlRequestRecord struct {
	OperationID, EnvironmentID, ActorUserID string
	IdempotencyKey                          string
	RequestSHA256                           []byte
	Kind, ClaimAlgorithm                    string
	ClaimPublicKey                          []byte
	OccurredAt                              time.Time
}

type RecoveryControlOperationRecord struct {
	OperationID, EnvironmentID, Kind string
	State, KitClaimState             string
	ArtifactSHA256, FailureCode      string
	ExpiresAt                        *time.Time
	CreatedAt                        time.Time
}

type RecoveryKitClaimRecord struct {
	ArtifactBytes  []byte
	ArtifactSHA256 string
	ExpiresAt      time.Time
}

type RecoveryProtectionStore interface {
	GetRecoveryProtection(context.Context, string) (RecoveryProtectionRecord, error)
	RequestRecoveryControl(
		context.Context, RecoveryControlRequestRecord,
	) (RecoveryControlOperationRecord, error)
	GetRecoveryControlOperation(
		context.Context, string,
	) (RecoveryControlOperationRecord, error)
	ClaimRecoveryKit(
		context.Context, string, string, string, time.Time,
	) (RecoveryKitClaimRecord, error)
	AcknowledgeRecoveryKit(
		context.Context, string, string, string, []byte, []byte, time.Time,
	) error
}
