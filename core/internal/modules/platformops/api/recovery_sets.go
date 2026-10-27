package api

import (
	"context"
	"time"
)

type DeploymentMode string

const (
	DeploymentOnPremise DeploymentMode = "onpremise"
	DeploymentSaaS      DeploymentMode = "saas"
)

type EnvironmentIdentity struct {
	EnvironmentID          string
	DeploymentMode         DeploymentMode
	PrimaryBindingID       string
	TenantSetSHA256        string
	OwnerBindingID         string
	OwnerBindingGeneration int64
	DatabaseLineageID      string
}

type RecoverySetRequest struct {
	RecoverySetID string
	OperationID   string
	EnvironmentID string
}

type RecoverySet struct {
	RecoverySetID    string
	OperationID      string
	EnvironmentID    string
	Sequence         int64
	State            string
	ManifestSHA256   string
	LegalHold        bool
	DatabaseSystemID string
	ProtectedAt      *time.Time
}

type RecoveryArtifact struct {
	SHA256 string
	Bytes  []byte
}

type RecoveryArtifactClaim struct {
	ArtifactID, TenantID, ActorID, SessionID string
}

type RetentionResult struct {
	Keep   []string
	Delete []string
}

type RecoveryRetentionCandidate struct {
	RecoverySetID  string
	ManifestSHA256 string
	ProtectedAt    time.Time
	LegalHold      bool
}

type RecoverySets interface {
	RequestRecoverySet(context.Context, RecoverySetRequest) (RecoverySet, error)
	GetRecoverySet(context.Context, string) (RecoverySet, error)
	ListRecoverySets(context.Context, string) ([]RecoverySet, error)
	ClaimRecoveryManifest(context.Context, RecoveryArtifactClaim) (RecoveryArtifact, error)
	MarkRecoverySetDeleting(context.Context, string, string) error
	PruneRecoverySets(context.Context, string) (RetentionResult, error)
}
