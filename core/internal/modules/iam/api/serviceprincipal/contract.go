package serviceprincipal

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Scope struct {
	Kind         string
	Owner        string
	ResourceKind string
	ResourceID   [16]byte
}

type VerificationRequest struct {
	PrincipalID    [16]byte
	GrantID        [16]byte
	GrantRevision  int64
	Capability     string
	Scope          Scope
	ManifestSHA256 [32]byte
	VerifyAt       time.Time
}

type VerificationEvidence struct {
	PrincipalID          [16]byte
	PrincipalRevision    int64
	AuthorizationVersion int64
	GrantID              [16]byte
	GrantRevision        int64
	Capability           string
	Scope                Scope
	ManifestSHA256       [32]byte
	VerifiedAt           time.Time
}

type EvidenceVerifier interface {
	VerifyServicePrincipalEvidence(
		context.Context,
		platformtx.TenantScope,
		VerificationRequest,
	) (VerificationEvidence, error)
}
