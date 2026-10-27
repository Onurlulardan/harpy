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

type PublicationAuthorityEvidence struct {
	ActorID                      [16]byte
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
	PurposeSHA256                [32]byte
	AuthenticatedAt              time.Time
	AMRPassword                  bool
	AMRWebAuthn                  bool
	AMRTOTP                      bool
}

type PublicationAuthorityRequest struct {
	DefinitionID     [16]byte
	ManifestSHA256   [32]byte
	IdempotencyKey   string
	MutationEvidence PublicationAuthorityEvidence
}

type PublicationAuthority struct {
	ServicePrincipalID [16]byte
	GrantID            [16]byte
	GrantRevision      int64
}

type PublicationAuthorityProvisioner interface {
	ProvisionWorkflowPublicationAuthority(
		context.Context,
		platformtx.TenantScope,
		PublicationAuthorityRequest,
	) (PublicationAuthority, error)
}
