package humantask

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Candidate struct {
	Kind string
	ID   [16]byte
}

type VerificationRequest struct {
	TenantID  [16]byte
	ActorID   [16]byte
	Candidate Candidate
	Action    string
	VerifyAt  time.Time
}

type VerificationEvidence struct {
	TenantID             [16]byte
	ActorID              [16]byte
	AuthorizationVersion int64
	Candidate            Candidate
	Action               string
	PermissionID         [16]byte
	EvidenceSHA256       [32]byte
	VerifiedAt           time.Time
}

type AssignmentRequest struct {
	TenantID   [16]byte
	Assignment Candidate
	VerifyAt   time.Time
}

type AssignmentEvidence struct {
	TenantID       [16]byte
	Assignment     Candidate
	Revision       int64
	EvidenceSHA256 [32]byte
	VerifiedAt     time.Time
}

type EvidenceVerifier interface {
	VerifyHumanTaskEvidence(
		context.Context,
		platformtx.TenantScope,
		VerificationRequest,
	) (VerificationEvidence, error)
	VerifyHumanTaskAssignment(
		context.Context,
		platformtx.TenantScope,
		AssignmentRequest,
	) (AssignmentEvidence, error)
}
