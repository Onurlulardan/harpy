package ports

import (
	"context"
	"time"
)

type RecoverySetRequestRecord struct {
	RecoverySetID string
	OperationID   string
	EnvironmentID string
	CreatedAt     time.Time
}

type RecoverySetRecord struct {
	RecoverySetID             string
	OperationID               string
	EnvironmentID             string
	Sequence                  int64
	State                     string
	ManifestSHA256            string
	ManifestBytes             []byte
	LegalHold                 bool
	DatabaseSystemID          string
	RecoverySigningGeneration int64
	RecoveryTrustBundleSHA256 string
	ProtectedAt               *time.Time
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type RecoveryArtifactRecord struct {
	SHA256 string
	Bytes  []byte
}

type RecoveryArtifactClaim struct {
	ArtifactID, TenantID, ActorID, SessionID string
	ClaimedAt                                time.Time
}

type RetentionRecord struct {
	RecoverySetID  string
	ManifestSHA256 string
	ProtectedAt    time.Time
	LegalHold      bool
}

type RecoverySetStore interface {
	RequestRecoverySet(context.Context, RecoverySetRequestRecord) (int64, error)
	GetRecoverySet(context.Context, string) (RecoverySetRecord, error)
	ListRecoverySets(context.Context, string) ([]RecoverySetRecord, error)
	ClaimRecoveryManifest(context.Context, RecoveryArtifactClaim) (RecoveryArtifactRecord, error)
	ListRetentionCandidates(context.Context, string) ([]RetentionRecord, error)
	MarkRecoverySetDeleting(context.Context, string, string, time.Time) error
}
