package iam

import (
	"context"

	serviceprincipalapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/serviceprincipal"
	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *ServicePrincipalRuntime) VerifyServicePrincipalEvidence(
	ctx context.Context,
	scope platformtx.TenantScope,
	request serviceprincipalapi.VerificationRequest,
) (serviceprincipalapi.VerificationEvidence, error) {
	if !runtime.Valid() || request.Validate() != nil {
		return serviceprincipalapi.VerificationEvidence{},
			serviceprincipalapi.ErrInvalidVerification
	}
	capability, err := serviceapp.ParseCapabilityKey(request.Capability)
	if err != nil {
		return serviceprincipalapi.VerificationEvidence{},
			serviceprincipalapi.ErrInvalidVerification
	}
	evidence, err := runtime.verifier.VerifyEvidenceAt(
		ctx, scope,
		serviceapp.PinnedRequest{
			PrincipalID: uuid.UUID(request.PrincipalID),
			GrantID:     uuid.UUID(request.GrantID), GrantRevision: request.GrantRevision,
			Capability: capability, Manifest: request.ManifestSHA256,
			Scope: serviceapp.ResourceScope{
				Kind: serviceapp.ScopeKind(request.Scope.Kind), Owner: request.Scope.Owner,
				ResourceKind: request.Scope.ResourceKind,
				ResourceID:   uuid.UUID(request.Scope.ResourceID),
			},
		}, request.VerifyAt,
	)
	if err != nil {
		return serviceprincipalapi.VerificationEvidence{}, err
	}
	result := serviceprincipalapi.VerificationEvidence{
		PrincipalID:          [16]byte(evidence.PrincipalID),
		PrincipalRevision:    evidence.PrincipalRevision,
		AuthorizationVersion: evidence.AuthorizationVersion,
		GrantID:              [16]byte(evidence.GrantID), GrantRevision: evidence.GrantRevision,
		Capability: string(evidence.Capability),
		Scope: serviceprincipalapi.Scope{
			Kind: string(evidence.Scope.Kind), Owner: evidence.Scope.Owner,
			ResourceKind: evidence.Scope.ResourceKind,
			ResourceID:   [16]byte(evidence.Scope.ResourceID),
		},
		ManifestSHA256: evidence.ManifestSHA256, VerifiedAt: evidence.VerifiedAt,
	}
	if !result.Matches(request) {
		return serviceprincipalapi.VerificationEvidence{},
			serviceprincipalapi.ErrInvalidVerification
	}
	return result, nil
}
