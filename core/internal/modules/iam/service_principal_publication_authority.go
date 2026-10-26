package iam

import (
	"context"
	"errors"

	serviceapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/serviceprincipal"
	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *ServicePrincipalRuntime) ProvisionWorkflowPublicationAuthority(
	ctx context.Context,
	scope platformtx.TenantScope,
	request serviceapi.PublicationAuthorityRequest,
) (serviceapi.PublicationAuthority, error) {
	if runtime == nil || runtime.authority == nil {
		return serviceapi.PublicationAuthority{}, serviceapp.ErrEvidenceMissing
	}
	result, err := runtime.authority.Provision(ctx, scope,
		serviceapp.ProvisionPublicationAuthority{
			DefinitionID:   uuid.UUID(request.DefinitionID),
			ManifestSHA256: request.ManifestSHA256,
			IdempotencyKey: request.IdempotencyKey,
			MutationEvidence: serviceapp.PublicationAuthorityEvidence{
				ActorID:                      uuid.UUID(request.MutationEvidence.ActorID),
				ExpectedTokenVersion:         request.MutationEvidence.ExpectedTokenVersion,
				ExpectedAuthorizationVersion: request.MutationEvidence.ExpectedAuthorizationVersion,
				PurposeSHA256:                request.MutationEvidence.PurposeSHA256,
				AuthenticatedAt:              request.MutationEvidence.AuthenticatedAt,
				AMRPassword:                  request.MutationEvidence.AMRPassword,
				AMRWebAuthn:                  request.MutationEvidence.AMRWebAuthn,
				AMRTOTP:                      request.MutationEvidence.AMRTOTP,
			},
		})
	if err != nil {
		return serviceapi.PublicationAuthority{}, publicationAuthorityError(err)
	}
	return serviceapi.PublicationAuthority{
		ServicePrincipalID: [16]byte(result.ServicePrincipalID),
		GrantID:            [16]byte(result.GrantID),
		GrantRevision:      result.GrantRevision,
	}, nil
}

var _ serviceapi.PublicationAuthorityProvisioner = (*ServicePrincipalRuntime)(nil)

func publicationAuthorityError(err error) error {
	if errors.Is(err, serviceapp.ErrGrantInactive) ||
		errors.Is(err, serviceapp.ErrGrantMismatch) {
		return serviceapi.ErrPublicationAuthorityConflict
	}
	return serviceapi.ErrPublicationAuthorityUnavailable
}
