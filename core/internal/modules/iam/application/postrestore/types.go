package postrestore

import (
	"time"

	"github.com/google/uuid"
)

const (
	StateReactivationRequired = "identity_reactivation_required"
	StateReactivated          = "reactivated"
	StateFenced               = "fenced"
)

type Admission struct {
	State              string
	EnvironmentID      uuid.UUID
	RestoreOperationID uuid.UUID
}

type Binding struct {
	TenantID            uuid.UUID
	OwnerUserID         uuid.UUID
	AnchorID            uuid.UUID
	NewDatabaseSystemID uuid.UUID
	SecurityEpoch       [32]byte
	TenantSlug          string
}

type ChallengeCommand struct {
	OperationID, EnvironmentID, RestoreOperationID uuid.UUID
	Challenge, RPID, Origin                        string
	ContextSHA256                                  [32]byte
	ExpiresAt, OccurredAt                          time.Time
}

type ChallengeRecord struct {
	OperationID, TenantID, OwnerUserID, AnchorID uuid.UUID
	Challenge, RPID                              string
	CredentialID, PublicKeyCOSE                  []byte
	ContextSHA256                                [32]byte
	SignCount                                    int64
	ExpiresAt                                    time.Time
	Replayed                                     bool
}

type OwnerIdentifier struct {
	ID               uuid.UUID
	NormalizedValue  string
	NormalizedSHA256 [32]byte
	Revision         int64
}

type ChallengeContext struct {
	ChallengeRecord
	Identifier OwnerIdentifier
}

type PasswordCredential struct {
	Salt        [16]byte
	Tag         [32]byte
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
}

type CompleteCommand struct {
	OperationID, EnvironmentID, RestoreOperationID uuid.UUID
	TenantID, OwnerUserID, AnchorID                uuid.UUID
	AssertionSHA256, PasswordSHA256, ReceiptSHA256 [32]byte
	SignCount                                      int64
	Password                                       PasswordCredential
	Identifier                                     OwnerIdentifier
	DirectoryKeyID                                 string
	DirectoryHMAC                                  [32]byte
	OccurredAt                                     time.Time
}

type ChallengeResult struct {
	OperationID uuid.UUID
	PublicKey   PublicKeyRequest
	ExpiresAt   time.Time
}

type PublicKeyRequest struct {
	Challenge        string                `json:"challenge"`
	RPID             string                `json:"rpId"`
	AllowCredentials []PublicKeyCredential `json:"allowCredentials"`
	UserVerification string                `json:"userVerification"`
	Timeout          int                   `json:"timeout"`
}

type PublicKeyCredential struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ReactivateResult struct {
	UserID, TenantID uuid.UUID
	TenantSlug       string
	Identifier       string
	SessionState     string
}

type ReconcileCommand struct {
	RestoreOperationID, EnvironmentID, RecoverySetID uuid.UUID
	TenantID, OwnerUserID, AnchorID                  uuid.UUID
	NewDatabaseSystemID                              uuid.UUID
	SecurityEpoch, RestoreReceiptSHA256              [32]byte
	ProtectedHeadSHA256, CustodyResultSHA256         [32]byte
	CustodyGeneration                                int64
	Receipt                                          []byte
	OccurredAt                                       time.Time
}

type ReconcileResult struct {
	State              string
	InvalidationCounts []byte
	InvalidationSHA256 [32]byte
	Replayed           bool
}
