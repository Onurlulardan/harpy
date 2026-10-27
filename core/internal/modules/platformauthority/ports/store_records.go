package ports

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

type OperationStatus string

const (
	StatusProposalSigned OperationStatus = "proposal_signed"
	StatusAckVerified    OperationStatus = "ack_verified"
	StatusCompleted      OperationStatus = "completed"
)

type OperationRecord struct {
	OperationID string
	Purpose     string
	State       artifact.CanonicalState
	Proposal    artifact.SignedProposal
	Ack         artifact.VerifiedAck
	Status      OperationStatus
	CompletedAt time.Time
}

type BindingRecord struct {
	BindingID                 string
	EnvironmentID             string
	DeploymentMode            artifact.DeploymentMode
	RegistryKind              artifact.RegistryKind
	LicenseBindingID          *string
	FinalizationReceiptSHA256 *string
	RecoveryTrustBundleSHA256 *string
	SignerGeneration          int64
	SignerKID                 string
	SignerPublicKeySHA256     string
	SignerStatus              string
}

type CommitResult struct {
	Operation OperationRecord
	Head      ExternalHead
}

type SignerRotationRecord struct {
	OperationID                  string
	BindingID                    string
	ExpectedSignerGeneration     int64
	ExpectedSignerKID            string
	NewSignerGeneration          int64
	NewSignerKID                 string
	NewSignerPublicKeySHA256     string
	NewRecoveryTrustBundleSHA256 *string
}

type SignerFenceRecord struct {
	BindingID                string
	ExpectedSignerGeneration int64
	ExpectedSignerKID        string
	EvidenceSHA256           string
}
