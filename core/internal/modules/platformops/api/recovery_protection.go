package api

import (
	"context"
	"time"
)

const RecoveryClaimAlgorithm = "HPKE-v1-X25519-HKDF-SHA256-AES-256-GCM"

type RecoveryProtection struct {
	EnvironmentID   string    `json:"environment_id"`
	State           string    `json:"state"`
	Generation      int64     `json:"generation"`
	KitAcknowledged bool      `json:"kit_acknowledged"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RecoveryControlRequest struct {
	OperationID, EnvironmentID, ActorUserID string
	IdempotencyKey, RequestSHA256           string
	Kind, ClaimAlgorithm                    string
	ClaimPublicKey                          []byte
}

type RecoveryControlOperation struct {
	OperationID, EnvironmentID, Kind string
	State, KitClaimState             string
	ArtifactSHA256, FailureCode      string
	ExpiresAt                        *time.Time
	CreatedAt                        time.Time
}

type RecoveryKitClaim struct {
	ArtifactBytes  []byte
	ArtifactSHA256 string
	ExpiresAt      time.Time
}

type RecoveryProtectionService interface {
	GetProtection(context.Context, string) (RecoveryProtection, error)
	RequestControl(context.Context, RecoveryControlRequest) (RecoveryControlOperation, error)
	GetControlOperation(context.Context, string) (RecoveryControlOperation, error)
	ClaimKit(context.Context, string, string, string) (RecoveryKitClaim, error)
	AcknowledgeKit(context.Context, string, string, string, string, []byte) error
}
